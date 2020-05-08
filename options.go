package testcontainers

import (
	"time"

	"github.com/romnnn/testcontainers-go"
)

// ContainerOptions ...
type ContainerOptions struct {
	testcontainers.ContainerRequest
	CollectLogs    bool
	StartupTimeout time.Duration
}

// ContainerConfig ...
type ContainerConfig struct {
	Log *LogCollector
}
