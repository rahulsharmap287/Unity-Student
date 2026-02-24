package models

import "time"

type ChatRequest struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Sender    string    `bson:"sender" json:"sender"`     // Jo request bhej raha hai
	Receiver  string    `bson:"receiver" json:"receiver"` // Jise request mil rahi hai
	Status    string    `bson:"status" json:"status"`     // "pending", "accepted", "rejected"
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
