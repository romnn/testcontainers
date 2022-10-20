package testcontainers

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
)

// ContainerOptions ...
type ContainerOptions struct {
	testcontainers.ContainerRequest
	StartupTimeout time.Duration
}

// ContainerConfig ...
type ContainerConfig struct {
}
