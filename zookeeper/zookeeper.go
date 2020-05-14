package zookeeper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	tc "github.com/romnnn/testcontainers"
	"github.com/romnnn/testcontainers-go"
	"github.com/romnnn/testcontainers-go/wait"
)

// ContainerOptions ...
type ContainerOptions struct {
	tc.ContainerOptions
}

// Config ...
type Config struct {
	tc.ContainerConfig
	Host string
	Port uint
	log  *tc.LogCollector
}

func (zkc Config) String() string {
	return fmt.Sprintf("%s:%d", zkc.Host, zkc.Port)
}

const defaultZookeeperPort = 2181

// StartZookeeperContainer ...
func StartZookeeperContainer(options ContainerOptions) (zkC testcontainers.Container, zkConfig *Config, err error) {
	ctx := context.Background()

	zookeeperPort, _ := nat.NewPort("", strconv.Itoa(defaultZookeeperPort))

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	// Do not expose any ports per default
	req := testcontainers.ContainerRequest{
		Image: "bitnami/zookeeper",
		Env: map[string]string{
			"ALLOW_ANONYMOUS_LOGIN": "yes",
		},
		WaitingFor: wait.ForLog("binding to port").WithStartupTimeout(timeout),
		/*
			Resources: &testcontainers.ContainerResourcers{
				Memory:     50 * 1024 * 1024, // max. 50MB
				MemorySwap: -1,               // Unlimited swap
			},
		*/
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	tc.ClientMux.Lock()
	zkC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	tc.ClientMux.Unlock()
	if err != nil {
		return
	}

	port := zookeeperPort
	if len(req.ExposedPorts) > 0 {
		port, err = zkC.MappedPort(ctx, zookeeperPort)
		if err != nil {
			return
		}
	}

	zkConfig = &Config{
		Host: "zookeeper",
		Port: uint(port.Int()),
	}

	if options.CollectLogs {
		zkConfig.ContainerConfig.Log = new(tc.LogCollector)
		go tc.EnableLogger(zkC, zkConfig.ContainerConfig.Log)
	}
	return
}
