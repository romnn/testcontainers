package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	tc "github.com/romnn/testcontainers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Options ...
type Options struct {
	tc.ContainerOptions
	Password string
	ImageTag string
}

// Container ...
type Container struct {
	Container testcontainers.Container
	tc.ContainerConfig
	Host     string
	Port     int64
	Password string
}

// Terminate ...
func (c *Container) Terminate(ctx context.Context) {
	if c.Container != nil {
		c.Container.Terminate(ctx)
	}
}

// ConnectionURI ...
func (c *Container) ConnectionURI() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Start...
func Start(ctx context.Context, options Options) (Container, error) {
	var container Container
	port, err := nat.NewPort("", "6379")
	if err != nil {
		return container, fmt.Errorf("failed to build port: %v", err)
	}

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	tag := "latest"
	if options.ImageTag != "" {
		tag = options.ImageTag
	}

	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("redis:%s", tag),
		ExposedPorts: []string{string(port)},
		WaitingFor:   wait.ForListeningPort(port).WithStartupTimeout(timeout),
	}

	if options.Password != "" {
		req.Cmd = []string{fmt.Sprintf("redis-server --requirepass %s", options.Password)}
		container.Password = options.Password
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	container.Container = redisContainer

	if err != nil {
		return container, fmt.Errorf("failed to start container: %v", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return container, fmt.Errorf("failed to get container host: %v", err)
	}
	container.Host = host

	realPort, err := redisContainer.MappedPort(ctx, port)
	if err != nil {
		return container, fmt.Errorf("failed to get exposed container port: %v", err)
	}
	container.Port = int64(realPort.Int())

	return container, nil
}
