package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

// DBClient is the mongoDB client that is used globally in the app
var DBClient *mongo.Client

// InitializeDB Initializes and return mongoDB client
func InitializeDB() (*mongo.Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://db:27017/url-shortener"))
	if err != nil { 
		return nil, err
	}
	return client, nil
}

// Insert : Inserts a new record in the url-shortener database
func Insert(db *mongo.Client, r Record) (Record, error) {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	res, err := db.Database("url-shortener").Collection("urls").InsertOne(ctx, r)
	
	if err != nil {
		return Record{}, err
	}
	log.Printf("New record created with the ID: %s", res.InsertedID)
	return r, nil
}

// GetByID : Search a record based on its id, and returns it
func GetByID(db *mongo.Client, ID string) (Record, error) {
	filter := bson.D{{Key: "id", Value: ID}}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	result := db.Database("url-shortener").Collection("urls").FindOne(ctx, filter)

	var record Record
	if result.Err() != nil {
		return record, result.Err()
	}
	result.Decode(&record)
	return record, nil
}

// FindByOriginalURL :  Search a record based on its originalURL, and returns it
func FindByOriginalURL(db *mongo.Client, url string) (Record, error) {
	filter := bson.D{{Key: "originalurl", Value: url}}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	result := db.Database("url-shortener").Collection("urls").FindOne(ctx, filter)

	var record Record
	if result.Err() != nil {
		return record, result.Err()
	}
	result.Decode(&record)
	return record, nil
}
