package main

import (
	"context"
	"log"
	"time"

	"ai-wrap/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(cfg.MongoDB.Database).Collection(cfg.MongoDB.Collection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "request_hash", Value: 1},
				{Key: "success", Value: 1},
			},
			Options: options.Index().SetName("cache_lookup"),
		},
		{
			Keys:    bson.D{{Key: "timestamp", Value: -1}},
			Options: options.Index().SetName("timestamp_desc"),
		},
		{
			Keys:    bson.D{{Key: "model", Value: 1}},
			Options: options.Index().SetName("model"),
		},
		{
			Keys:    bson.D{{Key: "success", Value: 1}},
			Options: options.Index().SetName("success"),
		},
		{
			Keys: bson.D{
				{Key: "timestamp", Value: -1},
				{Key: "success", Value: 1},
			},
			Options: options.Index().SetName("timestamp_success"),
		},
		{
			Keys: bson.D{
				{Key: "timestamp", Value: -1},
				{Key: "cache_hit", Value: 1},
			},
			Options: options.Index().SetName("timestamp_cache_hit"),
		},
	}

	log.Printf("creating indexes on %s.%s", cfg.MongoDB.Database, cfg.MongoDB.Collection)

	names, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Fatalf("failed to create indexes: %v", err)
	}

	log.Printf("created indexes: %v", names)
}
