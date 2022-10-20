package kafka

import (
	// "bufio"
	"context"
	"text/template"
	// "encoding/binary"
	"bytes"
	"fmt"
	// "io"
	"regexp"
	"strings"

	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	tc "github.com/romnn/testcontainers"
	tczk "github.com/romnn/testcontainers/zookeeper"
	"github.com/testcontainers/testcontainers-go"
)

// Options ...
type Options struct {
	tc.ContainerOptions
	LogLevel          string
	KafkaImageTag     string
	ZookeeperImageTag string
}

// Container ...
type KafkaContainer struct {
	tc.ContainerConfig
	Container testcontainers.Container
	// ZookeeperContainer *tczk.Container
	// Network            testcontainers.Network
	// tc.ContainerConfig
	Host    string
	Port    int
	Brokers []string
	Version string
}

// Container ...
type Container struct {
	Kafka *KafkaContainer
	// KafkaContainer testcontainers.Container
	// ZookeeperContainer testcontainers.Container
	Zookeeper *tczk.Container
	Network   testcontainers.Network
	// tc.ContainerConfig
	// Host         string
	// Port         int
	// Brokers      []string
	// KafkaVersion string
}

// // ZooContainer ...
// type Container struct {
// 	Container          testcontainers.Container
// 	ZookeeperContainer testcontainers.Container
// 	// ZookeeperContainer tczk.Container
// 	Network testcontainers.Network
// 	tc.ContainerConfig
// 	Host         string
// 	Port         int
// 	Brokers      []string
// 	KafkaVersion string
// }

// Terminate ...
func (c *KafkaContainer) Terminate(ctx context.Context) {
	if c.Container != nil {
		c.Container.Terminate(ctx)
	}
}

// Terminate ...
func (c *Container) Terminate(ctx context.Context) {
	if c.Kafka != nil {
		c.Kafka.Terminate(ctx)
	}

	if c.Zookeeper != nil {
		c.Zookeeper.Terminate(ctx)
	}

	if c.Network != nil {
		c.Network.Remove(ctx)
	}
}

// // Options ...
// type Options struct {
// 	tc.ContainerOptions
// 	LogLevel string
// 	ImageTag string
// }

// Start...
// kafkaC , Config *ContainerConnectionConfig, zkC testcontainers.Container, zkConfig *tczk.Config, net testcontainers.Network, err error
func Start(ctx context.Context, options Options) (Container, error) {
	var container Container
	port, err := nat.NewPort("", "9093")
	if err != nil {
		return container, fmt.Errorf("failed to build port: %v", err)
	}

	startScriptPath := "/start.sh"
	cmd := fmt.Sprintf("while [ ! -f %s ]; do sleep 0.1; done; cat %s && bash %s", startScriptPath, startScriptPath, startScriptPath)

	tag := "latest"
	if options.KafkaImageTag != "" {
		tag = options.KafkaImageTag
	}

	req := testcontainers.ContainerRequest{
		// Image: "confluentinc/cp-kafka:5.5.3",
		// Image: "confluentinc/cp-kafka:latest",
		Image: fmt.Sprintf("confluentinc/cp-kafka:%s", tag),
		Cmd:   []string{"/bin/bash", "-c", cmd},
		Env: map[string]string{
			"KAFKA_LISTENERS":                        fmt.Sprintf("PLAINTEXT://0.0.0.0:%d,BROKER://0.0.0.0:9092", port.Int()),
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":       "BROKER",
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":     "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS": "0",
		},
		Mounts: testcontainers.Mounts(testcontainers.BindMount("/var/run/docker.sock", "/var/run/docker.sock")),
		ExposedPorts: []string{
			string(port),
		},
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	// Create a network
	if len(req.Networks) < 1 {
		networkName := fmt.Sprintf("kafka-network-%s", tc.UniqueID())
		net, err := tc.CreateNetwork(testcontainers.NetworkRequest{
			Driver:         "bridge",
			Name:           networkName,
			Attachable:     true,
			CheckDuplicate: true,
		}, 2)
		if err != nil {
			return container, fmt.Errorf("failed to create network: %v", err)
		}
		req.Networks = []string{networkName}
		container.Network = net
	}

	// start zookeeper first
	aliases := make(map[string][]string)
	for _, net := range req.Networks {
		aliases[net] = []string{"zookeeper"}
	}
	zookeeperOptions := tczk.Options{
		ContainerOptions: tc.ContainerOptions{
			ContainerRequest: testcontainers.ContainerRequest{
				Networks:       req.Networks,
				NetworkAliases: aliases,
			},
		},
		ImageTag: options.ZookeeperImageTag,
	}
	zookeeperContainer, err := tczk.Start(ctx, zookeeperOptions)
	if err != nil {
		return container, fmt.Errorf("failed to start zookeeper container: %v", err)
	}
	container.Zookeeper = &zookeeperContainer

	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return container, fmt.Errorf("failed to start kafka container: %v", err)
	}
	container.Kafka = new(KafkaContainer)
	container.Kafka.Container = kafkaContainer

	host, err := kafkaContainer.Host(ctx)
	if err != nil {
		return container, fmt.Errorf("failed to get kafka container host: %v", err)
	}
	container.Kafka.Host = host

	realPort, err := kafkaContainer.MappedPort(ctx, port)
	if err != nil {
		return container, fmt.Errorf("failed to get exposed kafka container port: %v", err)
	}
	container.Kafka.Port = realPort.Int()
	container.Kafka.Brokers = []string{fmt.Sprintf("%s:%d", host, realPort.Int())}

	bootstrapServer := fmt.Sprintf("PLAINTEXT://%s:%d", host, realPort.Int())
	listeners := []string{bootstrapServer}

	client, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return container, fmt.Errorf("failed to get docker client: %v", err)
	}
	id := kafkaContainer.GetContainerID()
	inspect, err := client.ContainerInspect(context.Background(), id)
	if err != nil {
		return container, fmt.Errorf("failed to inspect container %s: %v", id, err)
	}
	for _, endpoint := range inspect.NetworkSettings.Networks {
		listener := fmt.Sprintf("BROKER://%s:9092", endpoint.IPAddress)
		listeners = append(listeners, listener)
	}

	logTemplatePath := "/etc/confluent/docker/log4j.properties.template.new"
	// logSettings := "log4j.logger.kafka=DEBUG,kafkaAppender"
	// logSettings := "log4j.logger.kafka=WARN,kafkaAppender"

	// log4j.rootLogger={{ env["KAFKA_LOG4J_ROOT_LOGLEVEL"] | default('INFO') }}, stdout
	// {% set loggers = {
	//   'kafka': 'INFO',
	//   'kafka.network.RequestChannel$': 'WARN',
	//   'kafka.producer.async.DefaultEventHandler': 'DEBUG',
	//   'kafka.request.logger': 'WARN',
	//   'kafka.controller': 'TRACE',
	//   'kafka.log.LogCleaner': 'INFO',
	//   'state.change.logger': 'TRACE',
	//   'kafka.authorizer.logger': 'WARN'
	//   } -%}
	// log4j.rootLogger=WARN, stdout
	// 'kafka': { env["KAFKA_LOG4J_ROOT_LOGLEVEL"] | default('INFO') },

	logTemplate := `
log4j.rootLogger={{ env["KAFKA_LOG4J_ROOT_LOGLEVEL"] | default('INFO') }}, stdout

log4j.appender.stdout=org.apache.log4j.ConsoleAppender
log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
log4j.appender.stdout.layout.ConversionPattern=[%d] %p %m (%c)%n
`

	// {% set loggers = {
	//   'kafka.network.RequestChannel$': 'WARN',
	//   'kafka.producer.async.DefaultEventHandler': 'DEBUG',
	//   'kafka.request.logger': 'WARN',
	//   'kafka.controller': 'TRACE',
	//   'kafka.log.LogCleaner': 'INFO',
	//   'state.change.logger': 'TRACE',
	//   'kafka.authorizer.logger': 'WARN'
	//   } -%}

	// {% if env['KAFKA_LOG4J_LOGGERS'] %}
	// {% set loggers = parse_log4j_loggers(env['KAFKA_LOG4J_LOGGERS'], loggers) %}
	// {% endif %}

	// {% for logger,loglevel in loggers.items() %}
	// log4j.logger.{{logger}}={{loglevel}}
	// {% endfor %}
	// `
	err = kafkaContainer.CopyToContainer(ctx, []byte(logTemplate), logTemplatePath, 700)
	if err != nil {
		return container, fmt.Errorf("failed to copy file to %s: %v", logTemplatePath, err)
	}

	versionCmd := []string{"kafka-topics", "--version"}
	versionOutput, err := tc.ExecCmd(kafkaContainer, versionCmd, ctx)
	if err != nil {
		return container, fmt.Errorf("failed to get kafka version: %v", err)
	}
	stdout := versionOutput.Stdout
	re := regexp.MustCompile(`^([\d.]*)-`)
	matches := re.FindStringSubmatch(stdout)
	if len(matches) != 2 {
		return container, fmt.Errorf(`failed to extract version from "%q"`, stdout)
	}
	container.Kafka.Version = matches[1]
	// version := matches[1]
	// log.Infof("kafka version: %s", version)

	// logCmd := []string{"/bin/bash", "-c", fmt.Sprintf(`export VAL="%s" && echo "$VAL" > %s && chmod 700 %s`, logTemplate, logProperties, logProperties)}
	// exitCode, reader, err := kafkaC.Exec(ctx, logCmd)
	// if err != nil || exitCode != 0 {
	// 	output, _ := ioutil.ReadAll(reader)
	// 	err = fmt.Errorf(`running: %s
	// in the kafka container failed:
	// exit code: %d
	// output: %s
	// error: %v`, logCmd, exitCode, output, err)
	// 	return
	// }

	// start script
	// export LOG_TEMPLATE=\"%s\"
	// echo "$LOG_TEMPLATE" > /etc/confluent/docker/log4j.properties.template
	// script := fmt.Sprintf(`#!/bin/bash
	scriptTmpl := template.Must(template.New("script").Parse(`#!/bin/bash

source /etc/confluent/docker/bash-config
export KAFKA_LOG4J_ROOT_LOGLEVEL="{{.LogLevel}}"
export KAFKA_ZOOKEEPER_CONNECT="{{.ZookeeperConnect}}"
export KAFKA_ADVERTISED_LISTENERS="{{.AdvListeners}}"
mv {{.LogTemplatePath}} /etc/confluent/docker/log4j.properties.template
/etc/confluent/docker/configure
/etc/confluent/docker/launch
`))

	logLevel := "WARN"
	if options.LogLevel != "" {
		logLevel = options.LogLevel
	}

	var script bytes.Buffer
	err = scriptTmpl.Execute(&script, struct {
		LogLevel         string
		ZookeeperConnect string
		AdvListeners     string
		LogTemplatePath  string
	}{
		LogLevel:         logLevel,
		ZookeeperConnect: fmt.Sprintf("%s:%d", zookeeperContainer.Host, zookeeperContainer.Port),
		AdvListeners:     strings.Join(listeners, ","),
		LogTemplatePath:  logTemplatePath,
	})
	if err != nil {
		return container, fmt.Errorf("failed to template start script: %v", err)
	}

	// script := fmt.Sprintf(`#!/bin/bash

	// source /etc/confluent/docker/bash-config
	// export KAFKA_LOG4J_ROOT_LOGLEVEL="%s"
	// export KAFKA_ZOOKEEPER_CONNECT="%s"
	// export KAFKA_ADVERTISED_LISTENERS="%s"
	// mv %s /etc/confluent/docker/log4j.properties.template
	// /etc/confluent/docker/configure
	// /etc/confluent/docker/launch
	// `, logLevel, zookeeperConnect, advListeners, logTemplatePath)

	err = kafkaContainer.CopyToContainer(ctx, script.Bytes(), startScriptPath, 700)
	if err != nil {
		return container, fmt.Errorf("failed to copy file to %s: %v", startScriptPath, err)
	}

	// strings.Replace(logTemplate, `"`, `\"`, -1)
	// script += fmt.Sprintf(`export KAFKA_ZOOKEEPER_CONNECT="%s:%d" && `, zkConfig.Host, zkConfig.Port)
	// script += fmt.Sprintf(`export KAFKA_ADVERTISED_LISTENERS="%s" && `, strings.Join(listeners, ","))
	// script += ". /etc/confluent/docker/bash-config && "
	// script += "/etc/confluent/docker/configure && "
	// script += "/etc/confluent/docker/launch"

	// startCmd := []string{"/bin/bash", "-c", fmt.Sprintf(`echo -e "%s" > %s && chmod 700 %s`, script, startScript, startScript)}
	// startCmd := []string{"/bin/bash", "-c", fmt.Sprintf(`echo -e "%s" > %s && chmod 700 %s`, script, startScript, startScript)}
	// _, err = exec(kafkaC, []string{"/bin/bash", "-c", startScriptPath}, ctx)
	// if err != nil {
	// 	return
	// }
	// exitCode, reader, err := kafkaC.Exec(ctx, startCmd)
	// if err != nil || exitCode != 0 {
	// 	output, _ := ioutil.ReadAll(reader)
	// 	err = fmt.Errorf(`running: %s
	// in the kafka container failed:
	// exit code: %d
	// output: %v
	// error: %v`, startCmd, exitCode, output, err)
	// 	return
	// }

	// Config = &ContainerConnectionConfig{
	// 	Brokers:      []string{fmt.Sprintf("%s:%d", host, realKafkaPort.Int())},
	// 	KafkaVersion: version,
	// }

	return container, nil
}
