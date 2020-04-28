package testcontainers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RabbitmqContainerOptions ...
type RabbitmqContainerOptions struct {
	ContainerOptions
}

// RabbitmqConfig ...
type RabbitmqConfig struct {
	ContainerConfig
	Host string
	Port int64
}

const (
	defaultRabbitmqPort = 5672
)

// StartRabbitmqContainer ...
func StartRabbitmqContainer(options RabbitmqContainerOptions) (rabbitmqC testcontainers.Container, rabbitmqConfig RabbitmqConfig, err error) {
	ctx := context.Background()
	rabbitmqPort, _ := nat.NewPort("", strconv.Itoa(defaultRabbitmqPort))

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.8.1-management",
		ExposedPorts: []string{string(rabbitmqPort)},
		WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(timeout),
	}

	mergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	clientMux.Lock()
	rabbitmqC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	clientMux.Unlock()
	if err != nil {
		err = fmt.Errorf("Failed to start rabbitmq container: %v", err)
		return
	}

	host, err := rabbitmqC.Host(ctx)
	if err != nil {
		err = fmt.Errorf("Failed to get rabbitmq container host: %v", err)
		return
	}

	port, err := rabbitmqC.MappedPort(ctx, rabbitmqPort)
	if err != nil {
		err = fmt.Errorf("Failed to get exposed rabbitmq container port: %v", err)
		return
	}

	rabbitmqConfig = RabbitmqConfig{
		Host: host,
		Port: int64(port.Int()),
	}

	if options.CollectLogs {
		rabbitmqConfig.ContainerConfig.Log = new(LogCollector)
		go enableLogger(rabbitmqC, rabbitmqConfig.ContainerConfig.Log)
	}
	return
}
