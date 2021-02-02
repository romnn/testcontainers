package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	tc "github.com/romnn/testcontainers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerOptions ...
type ContainerOptions struct {
	tc.ContainerOptions
	Password string
}

// Config ...
type Config struct {
	tc.ContainerConfig
	Host     string
	Port     int64
	Password string
}

// ConnectionURI ...
func (c Config) ConnectionURI() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

const (
	defaultRedisPort = 6379
)

// StartRedisContainer ...
func StartRedisContainer(ctx context.Context, options ContainerOptions) (redisC testcontainers.Container, redisConfig Config, err error) {
	redisPort, _ := nat.NewPort("", strconv.Itoa(defaultRedisPort))

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	req := testcontainers.ContainerRequest{
		Image:        "redis:6.0.10",
		ExposedPorts: []string{string(redisPort)},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(timeout),
	}

	if options.Password != "" {
		req.Cmd = []string{fmt.Sprintf("redis-server --requirepass %s", options.Password)}
		redisConfig.Password = options.Password
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	tc.ClientMux.Lock()
	redisC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	tc.ClientMux.Unlock()
	if err != nil {
		err = fmt.Errorf("Failed to start redis container: %v", err)
		return
	}

	host, err := redisC.Host(ctx)
	if err != nil {
		err = fmt.Errorf("Failed to get redis container host: %v", err)
		return
	}

	port, err := redisC.MappedPort(ctx, redisPort)
	if err != nil {
		err = fmt.Errorf("Failed to get exposed redis container port: %v", err)
		return
	}

	redisConfig.Host = host
	redisConfig.Port = int64(port.Int())

	if options.CollectLogs {
		redisConfig.ContainerConfig.Log = new(tc.LogCollector)
		go tc.EnableLogger(redisC, redisConfig.ContainerConfig.Log)
	}
	return
}
