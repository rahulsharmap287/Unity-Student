package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func SaveOTP(email, otp string) error {
	_, err := otpCollection.InsertOne(context.TODO(), bson.M{
		"email":     email,
		"otp":       otp,
		"expiresAt": time.Now().Add(5 * time.Minute),
		"verified":  false,
	})
	return err
}
