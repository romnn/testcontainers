package main

import (
	"context"
	"fmt"
	"time"

	tc "github.com/romnn/testcontainers"
	tcmongo "github.com/romnn/testcontainers/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func run() int {
	log.SetLevel(log.InfoLevel)
	// Start mongo container
	mongoC, mongoCfg, err := tcmongo.StartMongoContainer(context.Background(), tcmongo.ContainerOptions{
		ContainerOptions: tc.ContainerOptions{CollectLogs: true},
	})
	if err != nil {
		log.Fatalf("Failed to start mongoDB container: %v", err)
	}
	defer mongoC.Terminate(context.Background())
	var containerLog string
	if mongoCfg.Log != nil {
		go func() {
			for {
				msg := <-mongoCfg.Log.MessageChan
				containerLog = containerLog + msg
			}
		}()
		log.Infof("Collecting container logs for the next 15 seconds")
		time.Sleep(15 * time.Second)
		defer fmt.Println(containerLog)
	}

	// Connect to the database
	mongoURI := mongoCfg.ConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to create mongo client (%s): %v", mongoURI, err)
	}
	mctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client.Connect(mctx)

	err = client.Ping(mctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Could not ping database within %d seconds (%s): %v", 20, mongoURI, err)
	}
	database := client.Database("testdatabase")
	collection := database.Collection("my-collection")

	// Insert mock data
	insertMockData(collection)

	// Find adult users
	cur, err := collection.Find(context.Background(), bson.D{{Key: "age", Value: bson.D{{Key: "$gte", Value: 18}}}})
	if err != nil {
		log.Fatalf("Failed to query mongo database: %v", err)
	}
	defer cur.Close(context.Background())
	var numAdults int
	for cur.Next(context.Background()) {
		var result map[string]interface{}
		err := cur.Decode(&result)
		if err != nil {
			log.Fatalf("Failed to decode result: %v", err)
		}
		if !result["isAdult"].(bool) {
			log.Fatalf("Expected user to be adult, but got %v", result)
		}
		numAdults++
	}
	return numAdults
}

func insertMockData(collection *mongo.Collection) {
	var testdata []interface{}
	isAdult := func(age int) bool {
		if age >= 18 {
			return true
		}
		return false
	}
	for age := 0; age < 50; age++ {
		testdata = append(testdata, bson.D{{Key: "age", Value: age}, {Key: "isAdult", Value: isAdult(age)}})
	}
	_, err := collection.InsertMany(context.Background(), testdata)
	if err != nil {
		log.Fatalf("Failed to insert documents into mongo database: %v", err)
	}
}

func main() {
	log.Infof("Found %d adults in the database", run())
}
