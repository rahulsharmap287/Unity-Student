package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"Unity_Student/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
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
		"type":     "status",
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

		var message models.Message
		if err := json.Unmarshal(msg, &message); err != nil {
			continue
		}

		message.From = username
		message.CreatedAt = time.Now()

		// ✅ MongoDB mein ab ReplyToID aur ReplyText bhi save hoga
		res, _ := messageCollection.InsertOne(context.TODO(), message)

		// MongoDB ki InsertedID ko message mein wapas daalo taaki samne wala ise delete kar sake
		msgID := res.InsertedID.(primitive.ObjectID).Hex()

		mu.Lock()
		receiverConn, found := clients[message.To]
		mu.Unlock()

		if found {
			msgMap := map[string]interface{}{
				"type":      "chat_message",
				"messageId": msgID, // 👈 Message ID bhej rahe hain
				"from":      message.From,
				"text":      message.Text,
				"to":        message.To,
				"replyToId": message.ReplyToID, // 👈 Nayi Field: Reply ID
				"replyText": message.ReplyText, // 👈 Nayi Field: Reply Text
			}
			data, _ := json.Marshal(msgMap)
			receiverConn.WriteMessage(websocket.TextMessage, data)
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
