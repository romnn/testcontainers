package testcontainers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/prometheus/common/log"
	"github.com/testcontainers/testcontainers-go"
)

// KafkaContainerOptions ...
type KafkaContainerOptions struct {
	ContainerOptions
	StartZookeeper bool
}

// KafkaContainerConnectionConfig ...
type KafkaContainerConnectionConfig struct {
	ContainerConfig
	Brokers      []string
	KafkaVersion string
}

const (
	defaultKafkaPort = 9093
	startScript      = "/testcontainers_start.sh"
)

// StartStandaloneKafkaContainer ...
func StartStandaloneKafkaContainer(options KafkaContainerOptions) (testcontainers.Container, *KafkaContainerConnectionConfig, error) {
	options.StartZookeeper = false
	kafkaC, kafkaConfig, _, _, _, err := startKafkaContainer(options)
	return kafkaC, kafkaConfig, err
}

// StartKafkaContainer ...
func StartKafkaContainer(options KafkaContainerOptions) (testcontainers.Container, *KafkaContainerConnectionConfig, testcontainers.Container, testcontainers.Network, error) {
	options.StartZookeeper = true
	kafkaC, kafkaConfig, zkC, _, net, err := startKafkaContainer(options)
	return kafkaC, kafkaConfig, zkC, net, err
}

// StartKafkaContainer ...
func startKafkaContainer(options KafkaContainerOptions) (kafkaC testcontainers.Container, kafkaConfig *KafkaContainerConnectionConfig, zkC testcontainers.Container, zkConfig *ZookeeperConfig, net testcontainers.Network, err error) {
	ctx := context.Background()

	kafkaPort, _ := nat.NewPort("", strconv.Itoa(defaultKafkaPort))

	req := testcontainers.ContainerRequest{
		Image: "confluentinc/cp-kafka:5.5.0",
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

	mergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	// Create a network
	if len(req.Networks) < 1 {
		networkName := fmt.Sprintf("kafka-network-%s", UniqueID())
		createNetwork := func() error {
			var err error
			clientMux.Lock()
			net, err = testcontainers.GenericNetwork(context.Background(), testcontainers.GenericNetworkRequest{
				NetworkRequest: testcontainers.NetworkRequest{
					Driver:         "bridge",
					Name:           networkName,
					Attachable:     true,
					CheckDuplicate: true,
				},
			})
			clientMux.Unlock()
			return err
		}

		bo := backoff.NewExponentialBackOff()
		bo.Multiplier = 1.2
		bo.InitialInterval = 5 * time.Second
		bo.MaxElapsedTime = 2 * time.Minute
		err = backoff.Retry(createNetwork, bo)
		if err != nil {
			err = fmt.Errorf("Failed to create the docker test network: %v", err)
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
		zookeeperOptions := ZookeeperContainerOptions{
			ContainerOptions: ContainerOptions{
				ContainerRequest: testcontainers.ContainerRequest{
					Networks:       req.Networks,
					NetworkAliases: aliases,
				},
				CollectLogs: options.ContainerOptions.CollectLogs,
			},
		}
		zkC, zkConfig, err = StartZookeeperContainer(zookeeperOptions)
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

	kafkaConfig = &KafkaContainerConnectionConfig{
		Brokers:      []string{fmt.Sprintf("%s:%d", host, port.Int())},
		KafkaVersion: "5.5.0",
	}

	if options.CollectLogs {
		kafkaConfig.ContainerConfig.Log = new(LogCollector)
		go enableLogger(kafkaC, kafkaConfig.ContainerConfig.Log)
	}
	return
}
