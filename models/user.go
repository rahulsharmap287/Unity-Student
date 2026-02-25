package models

import "time"

type User struct {
	ID           string    `bson:"_id,omitempty"`
	Username     string    `bson:"username"`
	Email        string    `bson:"email"`
	CreatedAt    time.Time `bson:"createdAt"`
	ProfilePic   string    `bson:"profilePic" json:"profilePic"`
	College      string    `bson:"college" json:"college"` // ✨ EducationDetails se aayega
	Course       string    `bson:"course" json:"course"`   // ✨ EducationDetails se aayega
	BlockedUsers []string  `bson:"blockedUsers" json:"blockedUsers"`
	MutedUsers   []string  `bson:"mutedUsers" json:"mutedUsers"`
}
