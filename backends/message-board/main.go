package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Message struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type MessageStore struct {
	messages []Message
	mu       sync.RWMutex
	nextID   int
}

var store = &MessageStore{
	messages: make([]Message, 0),
	nextID:   1,
}

func (s *MessageStore) AddMessage(content string) Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:      s.nextID,
		Content: content,
	}
	s.messages = append(s.messages, msg)
	s.nextID++
	return msg
}

func (s *MessageStore) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.messages
}

func handlePostMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}

	message := store.AddMessage(request.Content)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	messages := store.GetMessages()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func main() {
	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlePostMessage(w, r)
		case http.MethodGet:
			handleGetMessages(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
