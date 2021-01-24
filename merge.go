package testcontainers

import (
	"github.com/imdario/mergo"
	"github.com/testcontainers/testcontainers-go"
)

// MergeRequest ...
func MergeRequest(c *testcontainers.ContainerRequest, override *testcontainers.ContainerRequest) {
	if err := mergo.Merge(c, override, mergo.WithOverride); err != nil {
		panic(err)
	}
}

// MergeOptions can merge generic options
func MergeOptions(c interface{}, override interface{}) {
	if err := mergo.Merge(c, override, mergo.WithOverride); err != nil {
		panic(err)
	}
}
