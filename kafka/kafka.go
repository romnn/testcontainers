package kafka

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"

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
type Container struct {
	tc.ContainerConfig
	Container testcontainers.Container
	Host      string
	Port      int
	Brokers   []string
	Listeners []string
	Version   string
}

// Composed ...
type Composed struct {
	Kafka     *Container
	Zookeeper *tczk.Container
	Network   testcontainers.Network
}

// Terminate ...
func (c *Container) Terminate(ctx context.Context) {
	if c.Container != nil {
		c.Container.Terminate(ctx)
	}
}

// Terminate ...
func (c *Composed) Terminate(ctx context.Context) {
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

func (c *Composed) getKafkaVersion(ctx context.Context) error {
	versionCmd := []string{"kafka-topics", "--version"}
	versionOutput, err := tc.ExecCmd(ctx, c.Kafka.Container, versionCmd)
	if err != nil {
		return fmt.Errorf("failed to get kafka version: %v", err)
	}
	stdout := versionOutput.Stdout
	re := regexp.MustCompile(`^([\d.]*)-`)
	matches := re.FindStringSubmatch(stdout)
	if len(matches) != 2 {
		return fmt.Errorf(`failed to extract version from "%q"`, stdout)
	}
	c.Kafka.Version = matches[1]
	return nil
}

func (c *Composed) addStartScript(ctx context.Context, path string, options Options) error {
	logTemplatePath := "/etc/confluent/docker/log4j.properties.template.new"
	logTemplate := `
log4j.rootLogger={{ env["KAFKA_LOG4J_ROOT_LOGLEVEL"] | default('INFO') }}, stdout

log4j.appender.stdout=org.apache.log4j.ConsoleAppender
log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
log4j.appender.stdout.layout.ConversionPattern=[%d] %p %m (%c)%n
`
	if err := c.Kafka.Container.CopyToContainer(ctx, []byte(logTemplate), logTemplatePath, 700); err != nil {
		return fmt.Errorf("failed to copy file to %s: %v", logTemplatePath, err)
	}

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

	zkHost := c.Zookeeper.Host
	zkPort := c.Zookeeper.Port
	var script bytes.Buffer
	err := scriptTmpl.Execute(&script, struct {
		LogLevel         string
		ZookeeperConnect string
		AdvListeners     string
		LogTemplatePath  string
	}{
		LogLevel:         logLevel,
		ZookeeperConnect: fmt.Sprintf("%s:%d", zkHost, zkPort),
		AdvListeners:     strings.Join(c.Kafka.Listeners, ","),
		LogTemplatePath:  logTemplatePath,
	})
	if err != nil {
		return fmt.Errorf("failed to template start script: %v", err)
	}

	if err := c.Kafka.Container.CopyToContainer(ctx, script.Bytes(), path, 700); err != nil {
		return fmt.Errorf("failed to copy file to %s: %v", path, err)
	}
	return nil
}

// Start ...
func Start(ctx context.Context, options Options) (Composed, error) {
	var composed Composed
	port, err := nat.NewPort("", "9093")
	if err != nil {
		return composed, fmt.Errorf("failed to build port: %v", err)
	}

	startScriptPath := "/start.sh"
	cmd := fmt.Sprintf("while [ ! -f %s ]; do sleep 0.1; done; cat %s && bash %s", startScriptPath, startScriptPath, startScriptPath)

	tag := "latest"
	if options.KafkaImageTag != "" {
		tag = options.KafkaImageTag
	}

	req := testcontainers.ContainerRequest{
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

	// create a network
	if len(req.Networks) < 1 {
		networkName := fmt.Sprintf("kafka-network-%s", tc.UniqueID())
		net, err := tc.CreateNetwork(testcontainers.NetworkRequest{
			Driver:         "bridge",
			Name:           networkName,
			Attachable:     true,
			CheckDuplicate: true,
		}, 2)
		if err != nil {
			return composed, fmt.Errorf("failed to create network: %v", err)
		}
		req.Networks = []string{networkName}
		composed.Network = net
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
		return composed, fmt.Errorf("failed to start zookeeper container: %v", err)
	}
	composed.Zookeeper = &zookeeperContainer

	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return composed, fmt.Errorf("failed to start kafka container: %v", err)
	}
	composed.Kafka = new(Container)
	composed.Kafka.Container = kafkaContainer

	host, err := kafkaContainer.Host(ctx)
	if err != nil {
		return composed, fmt.Errorf("failed to get kafka container host: %v", err)
	}
	composed.Kafka.Host = host

	realPort, err := kafkaContainer.MappedPort(ctx, port)
	if err != nil {
		return composed, fmt.Errorf("failed to get exposed kafka container port: %v", err)
	}
	composed.Kafka.Port = realPort.Int()
	composed.Kafka.Brokers = []string{fmt.Sprintf("%s:%d", host, realPort.Int())}

	bootstrapServer := fmt.Sprintf("PLAINTEXT://%s:%d", host, realPort.Int())
	composed.Kafka.Listeners = []string{bootstrapServer}

	client, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return composed, fmt.Errorf("failed to get docker client: %v", err)
	}
	id := kafkaContainer.GetContainerID()
	inspect, err := client.ContainerInspect(context.Background(), id)
	if err != nil {
		return composed, fmt.Errorf("failed to inspect container %s: %v", id, err)
	}
	for _, endpoint := range inspect.NetworkSettings.Networks {
		listener := fmt.Sprintf("BROKER://%s:9092", endpoint.IPAddress)
		composed.Kafka.Listeners = append(composed.Kafka.Listeners, listener)
	}

	err = composed.getKafkaVersion(ctx)
	if err != nil {
		return composed, err
	}

	err = composed.addStartScript(ctx, startScriptPath, options)
	if err != nil {
		return composed, err
	}

	return composed, nil
}
