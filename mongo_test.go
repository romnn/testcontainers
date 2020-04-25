package testcontainers

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// TestMongoContainer ...
func TestMongoContainer(t *testing.T) {
	t.Parallel()
	// Start mongo container
	mongoC, mongoConn, err := StartMongoContainer(MongoContainerOptions{})
	if err != nil {
		t.Fatalf("Failed to start mongoDB container: %v", err)
	}
	defer mongoC.Terminate(context.Background())

	// Connect to the database
	mongoURI := mongoConn.ConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("Failed to create mongo client (%s): %v", mongoURI, err)
	}
	mctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client.Connect(mctx)

	err = client.Ping(mctx, readpref.Primary())
	if err != nil {
		t.Fatalf("Could not ping database within %d seconds (%s): %v", 20, mongoURI, err)
	}
	database := client.Database("testdatabase")
	collection := database.Collection("my-collection")

	// Insert mock data
	insertedCount := 40
	if err := insertMockData(collection, insertedCount); err != nil {
		t.Fatalf("Failed to insert documents into mongo database: %v", err)
	}

	// Find all
	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		t.Fatalf("Failed to query mongo database: %v", err)
	}
	defer cur.Close(context.Background())
	var resultCount int
	for cur.Next(context.Background()) {
		var result map[string]interface{}
		err := cur.Decode(&result)
		if err != nil {
			t.Errorf("Failed to decode result: %v", err)
		}
		resultCount++
	}

	if resultCount != insertedCount {
		t.Fatalf("Expected %d items in the database but got %d", resultCount, insertedCount)
	}
}

func insertMockData(collection *mongo.Collection, num int) error {
	var testdata []interface{}
	for age := 0; age < num; age++ {
		testdata = append(testdata, bson.D{{Key: "age", Value: age}})
	}
	_, err := collection.InsertMany(context.Background(), testdata)
	return err
}
