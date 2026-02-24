package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Thought struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username   string             `json:"username" bson:"username"`
	Text       string             `json:"text" bson:"text"`
	ProfilePic string             `json:"profilePic" bson:"profilePic"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}
