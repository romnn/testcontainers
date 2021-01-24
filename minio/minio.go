package minio

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	tc "github.com/romnnn/testcontainers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerOptions ...
type ContainerOptions struct {
	tc.ContainerOptions
	AccessKeyID     string
	SecretAccessKey string
}

// Config ...
type Config struct {
	tc.ContainerConfig
	Host            string
	Port            uint
	AccessKeyID     string
	SecretAccessKey string
}

// ConnectionURI ...
func (c Config) ConnectionURI() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

const defaultMinioPort = 9000

// StartMinioContainer ...
func StartMinioContainer(ctx context.Context, options ContainerOptions) (minioC testcontainers.Container, config Config, err error) {
	minioPort, _ := nat.NewPort("", strconv.Itoa(defaultMinioPort))

	env := make(map[string]string)
	if options.AccessKeyID != "" && options.SecretAccessKey != "" {
		env["MINIO_ACCESS_KEY"] = options.AccessKeyID
		env["MINIO_SECRET_KEY"] = options.SecretAccessKey
	}

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:RELEASE.2021-01-16T02-19-44Z",
		Env:          env,
		Cmd:          []string{"server", "/data"},
		ExposedPorts: []string{string(minioPort)},
		WaitingFor:   wait.ForLog("Object API (Amazon S3 compatible)").WithStartupTimeout(timeout),
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	tc.ClientMux.Lock()
	minioC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	tc.ClientMux.Unlock()
	if err != nil {
		err = fmt.Errorf("Failed to start minio container: %v", err)
		return
	}

	host, err := minioC.Host(ctx)
	if err != nil {
		err = fmt.Errorf("Failed to get minio container host: %v", err)
		return
	}

	port, err := minioC.MappedPort(ctx, minioPort)
	if err != nil {
		err = fmt.Errorf("Failed to get exposed minio container port: %v", err)
		return
	}

	config = Config{
		Host:            host,
		Port:            uint(port.Int()),
		AccessKeyID:     options.AccessKeyID,
		SecretAccessKey: options.SecretAccessKey,
	}

	if options.CollectLogs {
		config.ContainerConfig.Log = &tc.LogCollector{
			MessageChan: make(chan string),
			Mux:         sync.Mutex{},
		}
		go tc.EnableLogger(minioC, config.ContainerConfig.Log)
	}
	return
}
