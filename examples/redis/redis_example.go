package main

import (
	"context"

	"github.com/go-redis/redis"
	tcredis "github.com/romnnn/testcontainers/redis"
	log "github.com/sirupsen/logrus"
)

func run() string {
	// Start redis container
	redisCont, redisConf, err := tcredis.StartRedisContainer(tcredis.RedisContainerOptions{})
	if err != nil {
		log.Fatalf("Failed to start redis container: %v", err)
	}
	defer redisCont.Terminate(context.Background())

	// Connect to redis database
	db := redis.NewClient(&redis.Options{
		Addr:     redisConf.ConnectionURI(),
		Password: redisConf.Password,
		DB:       1,
	})

	// Set some data
	db.HSet("my-hash-key", "key1", "Hello ")
	db.HSet("my-hash-key", "key2", "World!")

	// Get the data back
	k1, _ := db.HGet("my-hash-key", "key1").Result() // "Hello "
	k2, _ := db.HGet("my-hash-key", "key2").Result() // "World!"

	return k1 + k2
}

func main() {
	log.Infof("Received %s from redis", run())
}
