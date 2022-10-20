package zookeeper

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
	LogLevel string
	ImageTag string
}

// Container ...
type Container struct {
	Container testcontainers.Container
	tc.ContainerConfig
	Host string
	Port uint
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

// Start ...
func Start(ctx context.Context, options Options) (Container, error) {
	var container Container
	port, err := nat.NewPort("", "2181")
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

	logLevel := "WARN"
	if options.LogLevel != "" {
		logLevel = options.LogLevel
	}

	// Do not expose any ports per default
	req := testcontainers.ContainerRequest{
		Image: fmt.Sprintf("bitnami/zookeeper:%s", tag),
		Env: map[string]string{
			"ALLOW_ANONYMOUS_LOGIN": "yes",
			"ZOO_LOG_LEVEL":         logLevel,
		},
		WaitingFor: wait.ForListeningPort(port).WithStartupTimeout(timeout),
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	zookeeperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return container, fmt.Errorf("failed to start container: %v", err)
	}
	container.Container = zookeeperContainer

	realPort := port
	if len(req.ExposedPorts) > 0 {
		realPort, err = zookeeperContainer.MappedPort(ctx, port)
		if err != nil {
			return container, fmt.Errorf("failed to get exposed container port: %v", err)
		}

	}
	container.Port = uint(realPort.Int())
	container.Host = "zookeeper"

	return container, nil
}
