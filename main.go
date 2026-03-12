package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// 1. Port handle karein (Railway automatically PORT assign karta hai)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Local testing ke liye
	}

	ConnectMongo()

	fmt.Printf("✅ Unity Server started on port %s\n", port)

	http.HandleFunc("/send-otp", SendOTPHandler)
	http.HandleFunc("/send-request", SendRequestHandler)
	http.HandleFunc("/verify-otp", VerifyOTPHandler)
	http.HandleFunc("/check-username", CheckUsernameHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/users", GetAllUsers)
	http.HandleFunc("/new-users", GetNewUsersHandler)
	http.HandleFunc("/get-messages", GetChatHistoryHandler)
	http.HandleFunc("/get-accepted-friends", GetAcceptedFriendsHandler)
	http.HandleFunc("/search-users", SearchUsersHandler)
	http.HandleFunc("/markasread", MarkAsReadHandler)
	http.HandleFunc("/update-profile-pic", UpdateProfilePicHandler)
	http.HandleFunc("/block-user", BlockUserHandler)
	http.HandleFunc("/unblock-user", UnblockUserHandler)
	http.HandleFunc("/get-my-profile", GetMyProfileHandler)
	http.HandleFunc("/delete-message", DeleteMessageHandler)
	http.HandleFunc("/post-thought", PostThought)
	http.HandleFunc("/get-thoughts", GetThoughts)
	http.HandleFunc("/delete-thought", DeleteThought)
	http.HandleFunc("/get-user-profile", GetUserProfileDetails)
	http.HandleFunc("/get-notifications", GetNotifications)
	http.HandleFunc("/update-academic", UpdateAcademicHandler)
	http.HandleFunc("/get-user-thought", GetUserThoughtsHandler)
	http.HandleFunc("/toggle-mute", ToggleMuteHandler)
	http.HandleFunc("/delete-chat", DeleteChatHandler)
	http.HandleFunc("/send-notification", SendNotification)
	http.HandleFunc("/update-fcm-token", UpdateFCMToken)
	http.HandleFunc("/send-notification", SendNotification)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/ws", ServeWS)

	log.Printf("🚀 Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
