package testcontainers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ZookeeperContainerOptions ...
type ZookeeperContainerOptions struct {
	ContainerOptions
}

// ZookeeperConfig ...
type ZookeeperConfig struct {
	ContainerConfig
	Host string
	Port uint
	log  *LogCollector
}

func (zkc ZookeeperConfig) String() string {
	return fmt.Sprintf("%s:%d", zkc.Host, zkc.Port)
}

const defaultZookeeperPort = 2181

// StartZookeeperContainer ...
func StartZookeeperContainer(options ZookeeperContainerOptions) (zkC testcontainers.Container, zkConfig *ZookeeperConfig, err error) {
	ctx := context.Background()

	zookeeperPort, _ := nat.NewPort("", strconv.Itoa(defaultZookeeperPort))

	image := "bitnami/zookeeper"
	if options.ContainerOptions.Tag != "" {
		image += fmt.Sprintf(":%s", options.ContainerOptions.Tag)
	}

	// Do not expose any ports per default
	req := testcontainers.ContainerRequest{
		Image: image,
		Env: map[string]string{
			"ALLOW_ANONYMOUS_LOGIN": "yes",
		},
		WaitingFor: wait.ForLog("binding to port"),
	}

	mergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	clientMux.Lock()
	zkC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	clientMux.Unlock()
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

	zkConfig = &ZookeeperConfig{
		Host: "zookeeper",
		Port: uint(port.Int()),
	}

	if options.CollectLogs {
		zkConfig.ContainerConfig.Log = new(LogCollector)
		go enableLogger(zkC, zkConfig.ContainerConfig.Log)
	}
	return
}
