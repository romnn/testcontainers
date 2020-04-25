package testcontainers

import (
	"github.com/imdario/mergo"
	"github.com/testcontainers/testcontainers-go"
)

func mergeRequest(c *testcontainers.ContainerRequest, override *testcontainers.ContainerRequest) {
	if err := mergo.Merge(c, override, mergo.WithOverride); err != nil {
		panic(err)
	}
}
