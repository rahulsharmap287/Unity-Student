package models

import "time"

type Notification struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	ToUser    string    `bson:"toUser" json:"toUser"`     // Jise notification milegi
	FromUser  string    `bson:"fromUser" json:"fromUser"` // Jisne action kiya
	Type      string    `bson:"type" json:"type"`         // "message", "new_thought", "delete"
	Content   string    `bson:"content" json:"content"`   // Message ka chota part ya alert text
	IsRead    bool      `bson:"isRead" json:"isRead"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
