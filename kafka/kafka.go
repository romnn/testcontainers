package kafka

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/prometheus/common/log"
	tc "github.com/romnn/testcontainers"
	tczk "github.com/romnn/testcontainers/zookeeper"
	"github.com/testcontainers/testcontainers-go"
)

// ContainerOptions ...
type ContainerOptions struct {
	tc.ContainerOptions
	StartZookeeper bool
}

// ContainerConnectionConfig ...
type ContainerConnectionConfig struct {
	tc.ContainerConfig
	Brokers      []string
	KafkaVersion string
}

const (
	defaultKafkaPort = 9093
	startScript      = "/testcontainers_start.sh"
)

// StartStandaloneKafkaContainer ...
func StartStandaloneKafkaContainer(ctx context.Context, options ContainerOptions) (testcontainers.Container, *ContainerConnectionConfig, error) {
	options.StartZookeeper = false
	kafkaC, Config, _, _, _, err := startKafkaContainer(ctx, options)
	return kafkaC, Config, err
}

// StartKafkaContainer ...
func StartKafkaContainer(ctx context.Context, options ContainerOptions) (testcontainers.Container, *ContainerConnectionConfig, testcontainers.Container, testcontainers.Network, error) {
	options.StartZookeeper = true
	kafkaC, Config, zkC, _, net, err := startKafkaContainer(ctx, options)
	return kafkaC, Config, zkC, net, err
}

// StartKafkaContainer ...
func startKafkaContainer(ctx context.Context, options ContainerOptions) (kafkaC testcontainers.Container, Config *ContainerConnectionConfig, zkC testcontainers.Container, zkConfig *tczk.Config, net testcontainers.Network, err error) {
	kafkaPort, _ := nat.NewPort("", strconv.Itoa(defaultKafkaPort))

	req := testcontainers.ContainerRequest{
		Image: "confluentinc/cp-kafka:5.5.3",
		Cmd:   []string{"/bin/bash", "-c", fmt.Sprintf("while [ ! -f %s ]; do sleep 0.1; done; cat %s && %s", startScript, startScript, startScript)},
		Env: map[string]string{
			"KAFKA_LISTENERS":                        fmt.Sprintf("PLAINTEXT://0.0.0.0:%d,BROKER://0.0.0.0:9092", kafkaPort.Int()),
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":       "BROKER",
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":     "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS": "0",
		},
		ExposedPorts: []string{
			string(kafkaPort),
		},
		BindMounts: map[string]string{
			"/var/run/docker.sock": "/var/run/docker.sock",
		},
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	// Create a network
	if len(req.Networks) < 1 {
		networkName := fmt.Sprintf("kafka-network-%s", tc.UniqueID())
		net, err = tc.CreateNetwork(testcontainers.NetworkRequest{
			Driver:         "bridge",
			Name:           networkName,
			Attachable:     true,
			CheckDuplicate: true,
		}, 2)
		if err != nil {
			return
		}
		req.Networks = []string{networkName}
	}

	// Start zookeeper first
	if options.StartZookeeper {
		aliases := make(map[string][]string)
		for _, net := range req.Networks {
			aliases[net] = []string{"zookeeper"}
		}
		zookeeperOptions := tczk.ContainerOptions{
			ContainerOptions: tc.ContainerOptions{
				ContainerRequest: testcontainers.ContainerRequest{
					Networks:       req.Networks,
					NetworkAliases: aliases,
				},
				CollectLogs: options.ContainerOptions.CollectLogs,
			},
		}
		zkC, zkConfig, err = tczk.StartZookeeperContainer(ctx, zookeeperOptions)
		if err != nil {
			err = fmt.Errorf("Failed to start zookeeper container: %v", err)
			return
		}
	}

	kafkaC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return
	}

	host, err := kafkaC.Host(ctx)
	if err != nil {
		return
	}

	port, err := kafkaC.MappedPort(ctx, kafkaPort)
	if err != nil {
		return
	}

	bootstrapServer := fmt.Sprintf("PLAINTEXT://%s:%d", host, port.Int())
	listeners := []string{bootstrapServer}

	// Get docker client
	client, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return
	}
	inspect, err := client.ContainerInspect(context.Background(), kafkaC.GetContainerID())
	if err != nil {
		return
	}
	for _, endpoint := range inspect.NetworkSettings.Networks {
		listeners = append(listeners, fmt.Sprintf("BROKER://%s:9092", endpoint.IPAddress))
	}

	// Build script
	script := "#!/bin/bash \\n"
	script += fmt.Sprintf("export KAFKA_ZOOKEEPER_CONNECT=\"%s:%d\" && ", zkConfig.Host, zkConfig.Port)
	script += fmt.Sprintf("export KAFKA_ADVERTISED_LISTENERS=\"%s\" && ", strings.Join(listeners, ","))
	script += ". /etc/confluent/docker/bash-config && "
	script += "/etc/confluent/docker/configure && "
	script += "/etc/confluent/docker/launch"
	cmd := []string{"/bin/bash", "-c", fmt.Sprintf("printf '%s' > %s && chmod 700 %s", script, startScript, startScript)}
	log.Debug(cmd)

	exitCode, err := kafkaC.Exec(ctx, cmd)
	if err != nil {
		return
	}

	if exitCode != 0 {
		err = fmt.Errorf("Failed with code %d", exitCode)
		return
	}

	Config = &ContainerConnectionConfig{
		Brokers:      []string{fmt.Sprintf("%s:%d", host, port.Int())},
		KafkaVersion: "5.5.0",
	}

	if options.CollectLogs {
		Config.ContainerConfig.Log = new(tc.LogCollector)
		go tc.EnableLogger(kafkaC, Config.ContainerConfig.Log)
	}
	return
}
