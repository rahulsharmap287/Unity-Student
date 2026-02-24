package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive" // 👈 Ye import zaroori hai
)

type Message struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"` // 👈 Ye line add karni hai
	From      string             `json:"from" bson:"from"`
	To        string             `json:"to" bson:"to"`
	Text      string             `json:"text" bson:"text"`
	ReplyToID string             `json:"replyToId" bson:"replyToId,omitempty"` // Kis msg ka reply hai
	ReplyText string             `json:"replyText" bson:"replyText,omitempty"` // Us msg ka preview text
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
