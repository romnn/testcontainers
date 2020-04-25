package testcontainers

import "github.com/testcontainers/testcontainers-go"

// ContainerOptions ...
type ContainerOptions struct {
	testcontainers.ContainerRequest
	Tag         string
	CollectLogs bool
}

// ContainerConfig ...
type ContainerConfig struct {
	Log *LogCollector
}
