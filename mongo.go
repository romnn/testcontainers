package testcontainers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// MongoContainerOptions ...
type MongoContainerOptions struct {
	ContainerOptions
	User     string
	Password string
}

// MongoDBConfig ...
type MongoDBConfig struct {
	ContainerConfig
	Host     string
	Port     uint
	User     string
	Password string
}

// ConnectionURI ...
func (c MongoDBConfig) ConnectionURI() string {
	var databaseAuth string
	if c.User != "" && c.Password != "" {
		databaseAuth = fmt.Sprintf("%s:%s@", c.User, c.Password)
	}
	databaseHost := fmt.Sprintf("%s:%d", c.Host, c.Port)
	return fmt.Sprintf("mongodb://%s%s/?connect=direct", databaseAuth, databaseHost)
}

const defaultMongoDBPort = 27017

// StartMongoContainer ...
func StartMongoContainer(options MongoContainerOptions) (mongoC testcontainers.Container, mongoConfig MongoDBConfig, err error) {
	ctx := context.Background()
	mongoPort, _ := nat.NewPort("", strconv.Itoa(defaultMongoDBPort))

	var env map[string]string
	if options.User != "" && options.Password != "" {
		env["MONGO_INITDB_ROOT_USERNAME"] = options.User
		env["MONGO_INITDB_ROOT_PASSWORD"] = options.Password
	}

	image := "mongo"
	if options.ContainerOptions.Tag != "" {
		image += fmt.Sprintf(":%s", options.ContainerOptions.Tag)
	}

	req := testcontainers.ContainerRequest{
		Image:        image,
		Env:          env,
		ExposedPorts: []string{string(mongoPort)},
		WaitingFor:   wait.ForLog("waiting for connections on port"),
	}

	mergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	clientMux.Lock()
	mongoC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	clientMux.Unlock()
	if err != nil {
		err = fmt.Errorf("Failed to start mongo container: %v", err)
		return
	}

	host, err := mongoC.Host(ctx)
	if err != nil {
		err = fmt.Errorf("Failed to get mongo container host: %v", err)
		return
	}

	port, err := mongoC.MappedPort(ctx, mongoPort)
	if err != nil {
		err = fmt.Errorf("Failed to get exposed mongo container port: %v", err)
		return
	}

	mongoConfig = MongoDBConfig{
		Host:     host,
		Port:     uint(port.Int()),
		User:     options.User,
		Password: options.Password,
	}

	if options.CollectLogs {
		mongoConfig.ContainerConfig.Log = new(LogCollector)
		go enableLogger(mongoC, mongoConfig.ContainerConfig.Log)
	}
	return
}
