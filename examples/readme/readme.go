package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func testDatabaseFunction(mongoClient *mongo.Client) error {
	// Implement a real function here
	return nil
}

func main() {
	// Connect to the container
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/?connect=direct"))
	if err != nil {
		log.Fatalf("Failed to create mongo client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client.Connect(ctx)
	testDatabaseFunction(client)
}
