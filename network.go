package testcontainers

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/testcontainers/testcontainers-go"
)

// CreateNetwork creates a docker container network
func CreateNetwork(request testcontainers.NetworkRequest, timeoutMin time.Duration) (net testcontainers.Network, err error) {
	createNetwork := func() error {
		var err error
		ClientMux.Lock()
		net, err = testcontainers.GenericNetwork(context.Background(), testcontainers.GenericNetworkRequest{
			NetworkRequest: request,
		})
		ClientMux.Unlock()
		return err
	}

	bo := backoff.NewExponentialBackOff()
	bo.Multiplier = 1.2
	bo.InitialInterval = 5 * time.Second
	bo.MaxElapsedTime = timeoutMin * time.Minute
	err = backoff.Retry(createNetwork, bo)
	if err != nil {
		err = fmt.Errorf("Failed to create the docker test network: %v", err)
		return
	}
	return
}
