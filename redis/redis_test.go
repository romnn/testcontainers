package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// TestRabbitmqContainer ...
func TestRabbitmqContainer(t *testing.T) {
	t.Parallel()
	// Start redis container
	redisCont, redisConf, err := StartRedisContainer(ContainerOptions{})
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

	expected := "Hello World!"
	if k1+k2 != expected {
		t.Errorf("Expeceted to get %s from redis, but got %s", k1+k2, expected)
	}
}
