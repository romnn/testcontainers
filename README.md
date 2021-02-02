## testcontainers

[![GitHub](https://img.shields.io/github/license/romnn/testcontainers)](https://github.com/romnn/testcontainers)
[![GoDoc](https://godoc.org/github.com/romnn/testcontainers?status.svg)](https://godoc.org/github.com/romnn/testcontainers)  [![Test Coverage](https://codecov.io/gh/romnn/testcontainers/branch/master/graph/badge.svg)](https://codecov.io/gh/romnn/testcontainers)
[![Release](https://img.shields.io/github/release/romnn/testcontainers)](https://github.com/romnn/testcontainers/releases/latest)

A collection of pre-configured [testcontainers](https://github.com/testcontainers/testcontainers-go) for your golang integration tests.

Available containers (feel free to contribute):
- MongoDB (based on [mongo](https://hub.docker.com/_/mongo))
- Zookeeper (based on [bitnami/zookeeper](https://hub.docker.com/r/bitnami/zookeeper))
- Kafka (based on [wurstmeister/kafka](https://hub.docker.com/r/wurstmeister/kafka))
- RabbitMQ (based on [rabbitmq](https://hub.docker.com/_/rabbitmq))
- Redis (based on [redis](https://hub.docker.com/_/redis/))
- Minio (based on [minio/minio](https://hub.docker.com/r/minio/minio))

#### Usage

```golang
import (
	"context"
	"log"
	"time"

	tc "github.com/romnn/testcontainers/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func testDatabaseFunction(mongoClient *mongo.Client) error {
	_ = mongoClient.Database("testdatabase").Collection("my-collection")
	log.Println("Implement a real function here")
	return nil
}

func main() {
	// Start mongo container
	mongoC, mongoConn, err := tc.StartMongoContainer(context.Background(), tc.ContainerOptions{})
	if err != nil {
		log.Fatalf("Failed to start mongoDB container: %v", err)
	}
	defer mongoC.Terminate(context.Background())

	// Connect to the container
	mongoURI := mongoConn.ConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to create mongo client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client.Connect(ctx)
	testDatabaseFunction(client)
}
```

For full examples of all available containers, see `examples/`.


#### Development

######  Prerequisites

Before you get started, make sure you have installed the following tools::

    $ python3 -m pip install -U cookiecutter>=1.4.0
    $ python3 -m pip install pre-commit bump2version invoke ruamel.yaml halo
    $ go get -u golang.org/x/tools/cmd/goimports
    $ go get -u golang.org/x/lint/golint
    $ go get -u github.com/fzipp/gocyclo/cmd/gocyclo
    $ go get -u github.com/mitchellh/gox  # if you want to test building on different architectures

**Remember**: To be able to excecute the tools downloaded with `go get`, 
make sure to include `$GOPATH/bin` in your `$PATH`.
If `echo $GOPATH` does not give you a path make sure to run
(`export GOPATH="$HOME/go"` to set it). In order for your changes to persist, 
do not forget to add these to your shells `.bashrc`.

With the tools in place, it is strongly advised to install the git commit hooks to make sure checks are passing in CI:
```bash
invoke install-hooks
```

You can check if all checks pass at any time:
```bash
invoke pre-commit
```

Note for Maintainers: After merging changes, tag your commits with a new version and push to GitHub to create a release:
```bash
bump2version (major | minor | patch)
git push --follow-tags
```

#### Note

This project is still in the alpha stage and should not be considered production ready.
