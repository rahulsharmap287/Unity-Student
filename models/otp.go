package models

import "time"

type OTP struct {
	Email     string    `bson:"email"`
	OTP       string    `bson:"otp"`
	ExpiresAt time.Time `bson:"expiresAt"`
	Verified  bool      `bson:"verified"`
}
