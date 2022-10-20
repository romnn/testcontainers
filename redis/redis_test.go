package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redis"
	tc "github.com/romnn/testcontainers"
)

// TestRedis ...
func TestRedis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := Start(ctx, Options{
		ImageTag: "7.0.5",
	})
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}
	defer container.Terminate(ctx)

	// start logger
	logger, err := tc.StartLogger(ctx, container.Container)
	if err != nil {
		t.Errorf("failed to start logger: %v", err)
	} else {
		defer logger.Stop()
		go logger.LogToStdout()
	}

	// Connect to redis database
	db := redis.NewClient(&redis.Options{
		Addr:     container.ConnectionURI(),
		Password: container.Password,
		DB:       1,
	})

	// Set some data
	db.HSet("my-hash-key", "key1", "Hello")
	db.HSet("my-hash-key", "key2", "World!")

	// Get the data back
	k1, err := db.HGet("my-hash-key", "key1").Result() // "Hello"
	if err != nil {
		t.Errorf("failed to get key1: %v", err)
	}
	k2, err := db.HGet("my-hash-key", "key2").Result() // "World!"
	if err != nil {
		t.Errorf("failed to get key2: %v", err)
	}

	if k1 != "Hello" {
		t.Errorf(`expected "Hello", but got %s`, k1)
	}
	if k2 != "World!" {
		t.Errorf(`expected "World!", but got %s`, k2)
	}
}
