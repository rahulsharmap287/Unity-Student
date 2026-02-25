package main

import (
	"Unity_Student/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 1. Send OTP
func SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}

	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Email == "" {
		http.Error(w, "Email required", 400)
		return
	}
	otp := GenerateOTP()
	SaveOTP(body.Email, otp)
	SendOTPEmail(body.Email, otp)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent"})

}

func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	// 1. OTP Verify karein (Mongo se check)
	var record bson.M
	err := otpCollection.FindOne(context.TODO(), bson.M{"email": body.Email, "otp": body.OTP}).Decode(&record)
	if err != nil {
		http.Error(w, "Invalid OTP", 401)
		return
	}

	// 2. 🕵️ Check karein kya ye email pehle se 'userCollection' mein hai?
	var existingUser bson.M
	err = userCollection.FindOne(context.TODO(), bson.M{"email": body.Email}).Decode(&existingUser)

	if err == nil {
		// ✅ USER FOUND: Iska matlab ye purana user hai
		username := existingUser["username"].(string)
		token, _ := GenerateJWT(username) // Seedha token generate karo

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "old_user",
			"token":    token,
			"username": username,
			"message":  "Welcome back!",
		})
	} else {
		// 🆕 NEW USER: Iska email database mein nahi hai
		otpCollection.UpdateOne(context.TODO(), bson.M{"email": body.Email}, bson.M{"$set": bson.M{"verified": true}})
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "new_user",
			"message": "OTP verified, please complete registration",
		})
	}
}

// 3. Check Username
func CheckUsernameHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	count, _ := userCollection.CountDocuments(context.TODO(), bson.M{"username": body.Username})
	json.NewEncoder(w).Encode(map[string]bool{"available": count == 0})
}

// 4. Register User
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		FullName string `json:"fullName"`
		College  string `json:"college"`
		Course   string `json:"course"` // 👈 Step 1: Nayi field add ki
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}
	var otpRecord bson.M
	err := otpCollection.FindOne(context.TODO(), bson.M{"email": body.Email, "verified": true}).Decode(&otpRecord)
	if err != nil {
		http.Error(w, "Please verify your email with OTP first", 401)
		return
	}
	_, err = userCollection.InsertOne(context.TODO(), bson.M{
		"email":      body.Email,
		"username":   body.Username,
		"fullName":   body.FullName,
		"college":    body.College,
		"course":     body.Course, // 👈 Step 3: Yahan save ho rahi hai
		"profilePic": "",
		"createdAt":  time.Now(),
	})
	token, _ := GenerateJWT(body.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token, "username": body.Username})
}

// 5. Get All Users
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []bson.M
	cursor, _ := userCollection.Find(context.TODO(), bson.M{})
	cursor.All(context.TODO(), &users)
	json.NewEncoder(w).Encode(users)
}

// 6. Get New Users
func GetNewUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []bson.M
	cursor, err := userCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Mongo Error", 500)
		return
	}
	cursor.All(context.TODO(), &users)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// 7. Search Users
func SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	// Agar query khali hai toh empty list bhejien
	if query == "" {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	var users []bson.M
	// 🔍 Case-insensitive search: 'Rahul' search karne par 'rahul' bhi aayega
	filter := bson.M{
		"$or": []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"fullName": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := userCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	cursor.All(context.TODO(), &users)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// 8. Update Profile
func UpdateProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	json.NewDecoder(r.Body).Decode(&data)

	username := data["username"]
	profilePic := data["profilePic"]

	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"profilePic": profilePic}}

	_, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Update failed", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// 🔥 TELEGRAM STYLE: Auto-Accepted Request
func SendRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ChatRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Yahan status hamesha accepted rahega
	req.Status = "accepted"
	req.CreatedAt = time.Now()

	// Pehle check karein ki connection already hai ya nahi
	count, _ := requestCollection.CountDocuments(context.TODO(), bson.M{
		"$or": []bson.M{
			{"sender": req.Sender, "receiver": req.Receiver},
			{"sender": req.Receiver, "receiver": req.Sender},
		},
	})

	if count == 0 {
		_, _ = requestCollection.InsertOne(context.TODO(), req)
	}

	// Dono users ko WebSocket notify karein taaki list refresh ho jaye
	notifyRefresh := func(target string) {
		mu.Lock()
		conn, found := clients[target]
		mu.Unlock()
		if found {
			msg, _ := json.Marshal(map[string]string{"type": "new_chat"})
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}

	notifyRefresh(req.Sender)
	notifyRefresh(req.Receiver)

	json.NewEncoder(w).Encode(map[string]string{"message": "Chat initialized"})
}

func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// 🛡️ 1. Pehle check karein kya 'from' (main) ne 'to' (friend) ko block kiya hai?
	var user models.User
	_ = userCollection.FindOne(context.TODO(), bson.M{"username": from}).Decode(&user)

	isBlocked := false
	for _, b := range user.BlockedUsers {
		if b == to {
			isBlocked = true
			break
		}
	}

	var messages []models.Message
	filter := bson.M{
		"$or": []bson.M{
			{"from": from, "to": to},
			{"from": to, "to": from},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}})
	cursor, err := messageCollection.Find(context.TODO(), filter, opts)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	cursor.All(context.TODO(), &messages)

	// 🛡️ 2. Messages Filter logic: Agar blocked hai toh sirf mere bheje huye dikhao
	var filteredMessages []models.Message
	for _, msg := range messages {
		// Agar main 'from' hoon aur maine 'to' ko block kiya hai,
		// toh mujhe 'to' ke messages nahi dikhne chahiye
		if isBlocked && msg.From == to {
			continue // 🚫 Blocked user ka message skip karo
		}
		filteredMessages = append(filteredMessages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	// ✅ Ab filtered list bhej rahe hain
	if filteredMessages == nil {
		filteredMessages = []models.Message{} // Empty list safety
	}
	json.NewEncoder(w).Encode(filteredMessages)
}

func GetAcceptedFriendsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	var friendsData []map[string]interface{}

	var me models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&me)
	if err != nil {
		http.Error(w, "User not found", 404)
		return
	}

	filter := bson.M{
		"status": "accepted",
		"$or":    []bson.M{{"sender": username}, {"receiver": username}},
	}

	cursor, err := requestCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}

	var requests []models.ChatRequest
	cursor.All(context.TODO(), &requests)

	for _, req := range requests {
		friend := req.Receiver
		if req.Receiver == username {
			friend = req.Sender
		}

		// 🕵️ Sabse aakhri message nikalne ke liye (Sorting ke liye)
		var lastMsg models.Message
		msgFilter := bson.M{
			"$or": []bson.M{
				{"from": username, "to": friend},
				{"from": friend, "to": username},
			},
		}

		// 🕵️ Agar friend blocked hai, toh sirf mere bheje huye message dikhao (Subtitle mein)
		isBlocked := false
		for _, b := range me.BlockedUsers { // 'me' profile upar fetch kar lena
			if b == friend {
				isBlocked = true
				break
			}
		}

		if isBlocked {
			// Agar blocked hai, toh filter update karo taaki uske naye msgs count na ho
			msgFilter = bson.M{
				"from": username, // Sirf mere bheje huye
				"to":   friend,
			}
		}

		opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
		messageCollection.FindOne(context.TODO(), msgFilter, opts).Decode(&lastMsg)

		// 🕵️ Unread count (Jo humne pehle banaya tha)
		unreadCount := GetUnreadCount(username, friend)

		friendsData = append(friendsData, map[string]interface{}{
			"username":    friend,
			"lastMessage": lastMsg.Text,
			"time":        lastMsg.CreatedAt, // Isse Flutter sort karega
			"unreadCount": unreadCount,       // Isse green dot dikhega
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(friendsData)
}

// 1. Messages ko 'Read' mark karne ke liye
func MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	json.NewDecoder(r.Body).Decode(&data)

	myUsername := data["myUsername"]
	friendUsername := data["friendUsername"]

	// 🕵️ Filter: Wo saare messages jo friend ne mujhe bheje aur abhi tak unread hain
	filter := bson.M{"from": friendUsername, "to": myUsername, "isRead": false}
	update := bson.M{"$set": bson.M{"isRead": true}}

	_, err := messageCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Failed to update", 500)
		return
	}

	// 🔔 Notify the user via WebSocket to refresh the chat list (Optional but good)
	w.WriteHeader(http.StatusOK)
}

//  Function for Blocked user

func BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	json.NewDecoder(r.Body).Decode(&data)

	myUsername := data["myUsername"]
	targetUser := data["targetUser"]

	// 🕵️ 'myUsername' ke document mein 'blockedUsers' array mein 'targetUser' add karein
	filter := bson.M{"username": myUsername}
	update := bson.M{"$addToSet": bson.M{"blockedUsers": targetUser}}

	_, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Block failed", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Function for unblock user

func UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	json.NewDecoder(r.Body).Decode(&data)

	myUsername := data["myUsername"]
	targetUser := data["targetUser"]

	// 🕵️ 'blockedUsers' array se 'targetUser' ko nikalne ke liye $pull use karein
	filter := bson.M{"username": myUsername}
	update := bson.M{"$pull": bson.M{"blockedUsers": targetUser}}

	_, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Unblock failed", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", 404)
		return
	}

	// Sirf wahi data bhejien jo chahiye
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":     user.Username,
		"blockedUsers": user.BlockedUsers, // 👈 Ye array Flutter ko chahiye
	})
}

// 2. Unread count fetch karne ke liye (GetAcceptedFriendsHandler ke andar use hoga)
func GetUnreadCount(myUsername, friendUsername string) int64 {
	filter := bson.M{"from": friendUsername, "to": myUsername, "isRead": false}
	count, _ := messageCollection.CountDocuments(context.TODO(), filter)
	return count
}

func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", 400)
		return
	}

	messageID := data["messageId"]
	toUser := data["to"]

	// MongoDB ID mein convert karein
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		http.Error(w, "Invalid ID format", 400)
		return
	}

	// 🗑️ Database se delete karein
	_, err = messageCollection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}

	// 📢 Dusre user ko WebSocket signal bhejo
	NotifyDeleteToUser(toUser, messageID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// 1. Post Thought (Save to DB)
func PostThought(w http.ResponseWriter, r *http.Request) {
	var t models.Thought
	json.NewDecoder(r.Body).Decode(&t)
	t.CreatedAt = time.Now()

	_, err := thoughtCollection.InsertOne(context.TODO(), t)
	if err != nil {
		http.Error(w, "Failed to save thought", 500)
		return
		CreateNotification("all", t.Username, "new_thought", "posted a new thought")
	}
	w.WriteHeader(http.StatusOK)
}

// 2. Get All Thoughts (Fetch from DB)
func GetThoughts(w http.ResponseWriter, r *http.Request) {
	var thoughts []models.Thought

	// 🆕 Latest thoughts first dikhane ke liye sort option add kiya
	opts := options.Find().SetSort(bson.D{{"createdAt", -1}})

	cursor, err := thoughtCollection.Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		http.Error(w, "Error fetching thoughts", 500)
		return
	}
	cursor.All(context.TODO(), &thoughts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thoughts)
}

func GetUserProfileDetails(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	// 1. User ki details fetch karo
	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 2. Us user ke saare thoughts fetch karo
	// ✨ Logic: Initializing as empty slice to avoid null in JSON
	thoughts := []models.Thought{}

	// Sort logic add kiya taaki naye thoughts upar aayein
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := thoughtCollection.Find(context.TODO(), bson.M{"username": username}, opts)

	if err == nil {
		cursor.All(context.TODO(), &thoughts)
	}

	// 3. Response bhejo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":     user,
		"thoughts": thoughts, // 👈 Key 'thoughts' matches Flutter
	})
}

func DeleteThought(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Text     string `json:"text"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	// Database se match karke delete karo
	_, err := thoughtCollection.DeleteOne(context.TODO(), bson.M{
		"username": body.Username,
		"text":     body.Text,
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func CreateNotification(to, from, nType, content string) {
	notification := models.Notification{
		ToUser:    to,
		FromUser:  from,
		Type:      nType,
		Content:   content,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	// ✅ Ab ye asani se save ho jayega
	_, err := notificationCollection.InsertOne(context.TODO(), notification)
	if err != nil {
		fmt.Println("Notification Error:", err)
	}
}

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	// 1. Flutter se username lo (Query parameter se)
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// 2. Database se is user ki notifications dhundo
	var notifications []models.Notification
	// "toUser" match kar rahe hain kyunki notification usey milni hai
	cursor, err := notificationCollection.Find(context.TODO(), bson.M{"toUser": username})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// 3. Saara data slice mein bharo
	if err = cursor.All(context.TODO(), &notifications); err != nil {
		http.Error(w, "Error decoding data", http.StatusInternalServerError)
		return
	}

	// 4. JSON response bhejo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func UpdateAcademicHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Username string `json:"username"`
		FullName string `json:"fullName"` // ✨ Name bhi update hoga
		College  string `json:"college"`
		Course   string `json:"course"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid data", 400)
		return
	}

	// Database mein update query
	filter := bson.M{"username": data.Username}
	update := bson.M{"$set": bson.M{
		"fullName": data.FullName,
		"college":  data.College,
		"course":   data.Course,
	}}

	_, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile synced successfully"})
}

func GetUserThoughtsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", 400)
		return
	}

	filter := bson.M{"username": username}

	// ✨ models.Thought use karein bson.M ki jagah
	var thoughts []models.Thought

	cursor, err := thoughtCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	defer cursor.Close(context.TODO())

	// Data ko slice mein bharo
	if err = cursor.All(context.TODO(), &thoughts); err != nil {
		http.Error(w, "Decoding error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Ab ye small "text", "username" hi bhejega
	json.NewEncoder(w).Encode(thoughts)
}

func ToggleMuteHandler(w http.ResponseWriter, r *http.Request) {
	// Note: Flutter se hum 'username' bhej rahe hain userId ki jagah
	myUsername := r.URL.Query().Get("userId")
	targetUsername := r.URL.Query().Get("targetId")

	if myUsername == "" || targetUsername == "" {
		http.Error(w, "Missing usernames", http.StatusBadRequest)
		return
	}

	// 🕵️ Important: userCollection ka use karein jo pehle se defined hai
	// isse 'client' variable ka red error khatam ho jayega
	collection := userCollection

	// 1. Check karein kya targetUsername pehle se mutedUsers list mein hai
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"username": myUsername}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	isMuted := false
	for _, u := range user.MutedUsers {
		if u == targetUsername {
			isMuted = true
			break
		}
	}

	var update bson.M
	var message string

	if isMuted {
		// 2. UNMUTE logic
		update = bson.M{"$pull": bson.M{"mutedUsers": targetUsername}}
		message = "✅ User Unmuted"
	} else {
		// 3. MUTE logic
		update = bson.M{"$addToSet": bson.M{"mutedUsers": targetUsername}}
		message = "✅ User Muted"
	}

	_, err = collection.UpdateOne(context.TODO(), bson.M{"username": myUsername}, update)
	if err != nil {
		http.Error(w, "Database Update Failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, message)
}
