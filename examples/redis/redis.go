package main

import (
	"context"
	"log"

	"github.com/go-redis/redis"
	tc "github.com/romnn/testcontainers"
	tcredis "github.com/romnn/testcontainers/redis"
)

func main() {
	container, err := tcredis.Start(context.Background(), tcredis.Options{
		ImageTag: "7.0.5", // you could use latest here
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(context.Background())

	// start logger
	logger, err := tc.StartLogger(context.Background(), container.Container)
	if err != nil {
		log.Printf("failed to start logger: %v", err)
	} else {
		defer logger.Stop()
		go logger.LogToStdout()
	}

	// connect to redis
	db := redis.NewClient(&redis.Options{
		Addr:     container.ConnectionURI(),
		Password: container.Password,
		DB:       1,
	})

	// set some data
	db.HSet("my-hash-key", "key", "Hello World!")

	// get the data back
	value, err := db.HGet("my-hash-key", "key").Result()
	if err != nil {
		log.Fatalf("failed to get value: %v", err)
	}
	if value != "Hello World!" {
		log.Fatalf(`received %q instead of "Hello World!"`, value)
	}

	log.Printf("received %q from redis", value)
}
