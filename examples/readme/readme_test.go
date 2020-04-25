package main

import (
	"context"
	"testing"
	"time"

	tc "github.com/romnnn/testcontainers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestDatabaseIntegration tests datababase logic
func TestDatabaseIntegration(t *testing.T) {
	t.Parallel()
	// Start the container
	mongoC, mongoConfig, err := tc.StartMongoContainer(tc.MongoContainerOptions{})
	if err != nil {
		t.Fatalf("Failed to start mongoDB container: %v", err)
	}
	defer mongoC.Terminate(context.Background())

	// Connect to the container
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoConfig.ConnectionURI()))
	if err != nil {
		t.Fatalf("Failed to create mongo client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client.Connect(ctx)

	// TODO: Connect to the database using the URI and implement testDatabaseFunction!
	if err := testDatabaseFunction(client); err != nil {
		t.Fatalf("myDatabaseFunction failed with error: %v", err)
	}
}
