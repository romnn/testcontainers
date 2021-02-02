package main

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
