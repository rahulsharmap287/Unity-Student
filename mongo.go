package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variables taaki handler.go mein use ho sakein
var mongoClient *mongo.Client
var messageCollection *mongo.Collection
var otpCollection *mongo.Collection
var userCollection *mongo.Collection
var requestCollection *mongo.Collection
var thoughtCollection *mongo.Collection // 👈 Global rakhein
var notificationCollection *mongo.Collection

func ConnectMongo() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("❌ MONGO_URI not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	mongoClient = client
	db := mongoClient.Database("Unity_Student")

	// Collection Initialization
	messageCollection = db.Collection("messages")
	otpCollection = db.Collection("otp")
	userCollection = db.Collection("users")
	thoughtCollection = db.Collection("thoughts") // 👈 Isse index lagane mein madad milegi
	notificationCollection = db.Collection("notifications")
	requestCollection = db.Collection("chat_requests")

	// ✨ USERNAME INDEX: Search speed badhane ke liye
	_, err = userCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// ✨ THOUGHTS INDEX: Specific user thoughts jaldi load karne ke liye
	_, _ = thoughtCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{Key: "username", Value: 1}},
	})

	log.Println("✅ MongoDB connected & Optimized Indexes set")
}
