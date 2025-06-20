package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cstevenson98/goFE/pkg/shared"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// In-memory storage for demo purposes
var users = make(map[string]shared.User)
var messages = make(map[string]shared.Message)

func main() {
	// Initialize with some sample data
	initializeSampleData()

	r := mux.NewRouter()

	// User routes
	r.HandleFunc("/api/users", getUsers).Methods("GET")
	r.HandleFunc("/api/users", createUser).Methods("POST")
	r.HandleFunc("/api/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/api/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/api/users/{id}", deleteUser).Methods("DELETE")

	// Message routes
	r.HandleFunc("/api/messages", getMessages).Methods("GET")
	r.HandleFunc("/api/messages", createMessage).Methods("POST")
	r.HandleFunc("/api/messages/{id}", getMessage).Methods("GET")
	r.HandleFunc("/api/messages/{id}", deleteMessage).Methods("DELETE")

	// CORS middleware
	r.Use(corsMiddleware)

	fmt.Println("API server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func initializeSampleData() {
	// Create sample users
	user1 := shared.User{
		ID:        uuid.New().String(),
		Email:     "john@example.com",
		Name:      "John Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users[user1.ID] = user1

	user2 := shared.User{
		ID:        uuid.New().String(),
		Email:     "jane@example.com",
		Name:      "Jane Smith",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users[user2.ID] = user2

	// Create sample messages
	message1 := shared.Message{
		ID:        uuid.New().String(),
		Content:   "Hello, this is a test message!",
		UserID:    user1.ID,
		CreatedAt: time.Now(),
	}
	messages[message1.ID] = message1

	message2 := shared.Message{
		ID:        uuid.New().String(),
		Content:   "Another test message from Jane",
		UserID:    user2.ID,
		CreatedAt: time.Now(),
	}
	messages[message2.ID] = message2
}

// User handlers

func getUsers(w http.ResponseWriter, r *http.Request) {
	userList := make([]shared.User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}

	response := shared.APIResponse[shared.UsersResponse]{
		Data: shared.UsersResponse{
			Users: userList,
			Total: len(userList),
		},
		Success: true,
		Message: "Users retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var request shared.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Email == "" || request.Name == "" {
		http.Error(w, "Email and name are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	for _, user := range users {
		if user.Email == request.Email {
			http.Error(w, "User with this email already exists", http.StatusConflict)
			return
		}
	}

	// Create new user
	newUser := shared.User{
		ID:        uuid.New().String(),
		Email:     request.Email,
		Name:      request.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users[newUser.ID] = newUser

	response := shared.APIResponse[shared.UserResponse]{
		Data: shared.UserResponse{
			User: newUser,
		},
		Success: true,
		Message: "User created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, exists := users[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := shared.APIResponse[shared.UserResponse]{
		Data: shared.UserResponse{
			User: user,
		},
		Success: true,
		Message: "User retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, exists := users[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var request shared.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update user
	user.Name = request.Name
	user.UpdatedAt = time.Now()
	users[userID] = user

	response := shared.APIResponse[shared.UserResponse]{
		Data: shared.UserResponse{
			User: user,
		},
		Success: true,
		Message: "User updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if _, exists := users[userID]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Delete user
	delete(users, userID)

	// Delete associated messages
	for messageID, message := range messages {
		if message.UserID == userID {
			delete(messages, messageID)
		}
	}

	response := shared.APIResponse[map[string]interface{}]{
		Data:    map[string]interface{}{"id": userID},
		Success: true,
		Message: "User deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Message handlers

func getMessages(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	messageList := make([]shared.Message, 0, len(messages))
	for _, message := range messages {
		messageList = append(messageList, message)
	}

	// Simple pagination (in a real app, you'd use a database)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(messageList) {
		start = len(messageList)
	}
	if end > len(messageList) {
		end = len(messageList)
	}

	var paginatedMessages []shared.Message
	if start < len(messageList) {
		paginatedMessages = messageList[start:end]
	}

	response := shared.APIResponse[shared.MessagesResponse]{
		Data: shared.MessagesResponse{
			Messages: paginatedMessages,
			Total:    len(messageList),
		},
		Success: true,
		Message: "Messages retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createMessage(w http.ResponseWriter, r *http.Request) {
	var request shared.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Content == "" || request.UserID == "" {
		http.Error(w, "Content and user_id are required", http.StatusBadRequest)
		return
	}

	// Check if user exists
	if _, exists := users[request.UserID]; !exists {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Create new message
	newMessage := shared.Message{
		ID:        uuid.New().String(),
		Content:   request.Content,
		UserID:    request.UserID,
		CreatedAt: time.Now(),
	}
	messages[newMessage.ID] = newMessage

	response := shared.APIResponse[shared.MessageResponse]{
		Data: shared.MessageResponse{
			Message: newMessage,
		},
		Success: true,
		Message: "Message created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	message, exists := messages[messageID]
	if !exists {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	response := shared.APIResponse[shared.MessageResponse]{
		Data: shared.MessageResponse{
			Message: message,
		},
		Success: true,
		Message: "Message retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	if _, exists := messages[messageID]; !exists {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Delete message
	delete(messages, messageID)

	response := shared.APIResponse[map[string]interface{}]{
		Data:    map[string]interface{}{"id": messageID},
		Success: true,
		Message: "Message deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
