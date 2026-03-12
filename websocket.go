package main

import (
	"Unity_Student/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var clients = make(map[string]*websocket.Conn)
var mu sync.Mutex

// var thoughtCollection *mongo.Collection // 👈 Ye line add karein
// 🔹 Naya Function: Sabko status batane ke liye
func broadcastStatus(username string, isOnline bool) {
	mu.Lock()
	defer mu.Unlock()

	statusMsg := map[string]interface{}{
		"type": "status",

		"username": username,
		"online":   isOnline,
	}
	data, _ := json.Marshal(statusMsg)

	for _, conn := range clients {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

func ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Token missing", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	username, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "Identification failed", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	mu.Lock()
	clients[username] = conn
	mu.Unlock()

	log.Printf("✅ User Online: %s", username)

	// 🟢 Status Broadcast: Online
	broadcastStatus(username, true)

	defer func() {
		mu.Lock()
		delete(clients, username)
		mu.Unlock()

		// 🔴 Status Broadcast: Offline
		broadcastStatus(username, false)

		conn.Close()
		log.Printf("❌ User Offline: %s", username)
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 1. Pehle raw map mein parse karein taaki 'type' check kar sakein
		var rawData map[string]interface{}
		if err := json.Unmarshal(msg, &rawData); err != nil {
			continue
		}

		// 2. Sirf tabhi save karein agar ye chat message hai
		if rawData["type"] == "chat_message" {
			var message models.Message

			// Map se Message struct mein data bharein
			message.From = username
			message.To = fmt.Sprintf("%v", rawData["to"])
			message.Text = fmt.Sprintf("%v", rawData["text"])
			message.ReplyToID = fmt.Sprintf("%v", rawData["replyToId"])
			message.ReplyText = fmt.Sprintf("%v", rawData["replyText"])
			message.CreatedAt = time.Now()

			// 3. MongoDB mein Save karein
			res, err := messageCollection.InsertOne(context.TODO(), message)
			if err != nil {
				fmt.Println("❌ Mongo Save Error:", err)
				continue
			}

			// 4. Message ID nikal kar receiver ko bhejien
			msgID := res.InsertedID.(primitive.ObjectID).Hex()

			mu.Lock()
			receiverConn, found := clients[message.To]
			mu.Unlock()

			if found {
				msgMap := map[string]interface{}{
					"type":      "chat_message",
					"messageId": msgID,
					"from":      message.From,
					"text":      message.Text,
					"to":        message.To,
					"replyToId": message.ReplyToID,
					"replyText": message.ReplyText,
				}
				data, _ := json.Marshal(msgMap)
				receiverConn.WriteMessage(websocket.TextMessage, data)
			} else {
				// 🔴 Receiver Offline hai: Firebase Push Notification bhejo
				go func(receiverName string, senderName string, msgText string) {
					var receiver models.User
					// Database se receiver ka FCM Token nikalen
					err := userCollection.FindOne(context.TODO(), bson.M{"username": receiverName}).Decode(&receiver)

					if err == nil && receiver.FCMToken != "" {
						title := fmt.Sprintf("New message from %s", senderName)
						// Aapka helper function call karein
						TriggerPushNotification(receiver.FCMToken, title, msgText, message.From)
					}
				}(message.To, message.From, message.Text)
			}
		}
	}

}

func NotifyDeleteToUser(toUser string, msgID string) {
	mu.Lock()
	conn, found := clients[toUser]
	mu.Unlock()

	if found {
		// Signal bhejo jiska type 'delete_message' ho
		deleteSignal := map[string]interface{}{
			"type":      "delete_message",
			"messageId": msgID,
		}
		data, _ := json.Marshal(deleteSignal)
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}
