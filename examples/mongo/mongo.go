package main

import (
	"context"
	"log"
	"time"

	tcmongo "github.com/romnn/testcontainers/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	container, err := tcmongo.Start(context.Background(), tcmongo.Options{
		ImageTag: "6.0.2", // you could use latest here
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(context.Background())

	// connect to the container
	uri := container.ConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.Connect(ctx)

	// count documents in collection
	collection := client.Database("testdatabase").Collection("my-collection")
	opts := options.Count().SetMaxTime(2 * time.Second)
	count, err := collection.CountDocuments(
		context.TODO(),
		bson.D{},
		opts,
	)
	if err != nil {
		log.Fatalf("failed to count docs in collection %q: %v", collection.Name(), err)
	}
	log.Printf("collection %q contains %d documents", collection.Name(), count)
}
