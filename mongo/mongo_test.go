package mongo

import (
	"context"
	"testing"
	"time"

	tc "github.com/romnn/testcontainers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// TestMongo ...
func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := Start(ctx, Options{
    ImageTag: "6.0.2",
  })
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
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

	// connect to the database
	mongoURI := container.ConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("failed to create client (%s): %v", mongoURI, err)
	}
	mctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	client.Connect(mctx)

	err = client.Ping(mctx, readpref.Primary())
	if err != nil {
		t.Fatalf("could not ping %s within %d seconds: %v", mongoURI, 20, err)
	}
	database := client.Database("testdatabase")
	collection := database.Collection("my-collection")

	// insert mock data
	insertedCount := 40
	if err := insertMockData(ctx, collection, insertedCount); err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// find all
	cur, err := collection.Find(ctx, bson.D{{}})
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	defer cur.Close(ctx)
	var resultCount int
	for cur.Next(ctx) {
		var result map[string]interface{}
		err := cur.Decode(&result)
		if err != nil {
			t.Errorf("failed to decode result: %v", err)
		}
		resultCount++
	}

	if resultCount != insertedCount {
		t.Fatalf("expected %d documents but got %d", resultCount, insertedCount)
	}
}

func insertMockData(ctx context.Context, collection *mongo.Collection, num int) error {
	var testdata []interface{}
	for age := 0; age < num; age++ {
		testdata = append(testdata, bson.D{{Key: "age", Value: age}})
	}
	_, err := collection.InsertMany(ctx, testdata)
	return err
}
