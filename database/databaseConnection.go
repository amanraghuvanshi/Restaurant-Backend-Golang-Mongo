package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This function is responsible for connecting the database
// Will connect the database to running on 27107
func DBInstance() *mongo.Client {
	MongoDb := "mongodb://localhost:27107"
	fmt.Println(MongoDb)

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDb))

	if err != nil {
		log.Fatal("Error, while connecting")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.Connect(ctx); err != nil {
		log.Fatal("Error while connecting")
	}
	fmt.Println("Connected to MongoDB")
	return client
}

var Client *mongo.Client = DBInstance()

// Initializing the database for the restaurant, this will be accessing the database if exist, otherwise will create one.
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = (*mongo.Collection)(client.Database("restaurant").Collection(collectionName))

	return collection
}
