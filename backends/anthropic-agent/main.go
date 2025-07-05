package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/agent"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/types"
	"github.com/gorilla/mux"
)

// Global agent instance
var agentInstance agent.AnthropicAgent

func main() {
	// Initialize the Anthropic agent
	agentInstance = agent.NewAnthropicAgent()
	if err := agentInstance.Instantiate(); err != nil {
		log.Fatalf("Failed to initialize Anthropic agent: %v", err)
	}

	r := mux.NewRouter()

	// Apply CORS middleware first
	r.Use(corsMiddleware)

	// Anthropic agent routes
	r.HandleFunc("/api/chat", handleChat).Methods("POST", "OPTIONS")

	// Health check endpoint
	r.HandleFunc("/health", healthCheck).Methods("GET")

	// Endpoints documentation
	r.HandleFunc("/endpoints", getEndpoints).Methods("GET")

	fmt.Println("Anthropic Agent API server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
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

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request types.AnthropicRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Send message to Anthropic agent
	response, err := agentInstance.SendMessage(request.Message)
	if err != nil {
		apiResponse := types.APIResponse[map[string]interface{}]{
			Data:    map[string]interface{}{},
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiResponse)
		return
	}

	// Create response
	apiResponse := types.APIResponse[types.AnthropicResponse]{
		Data: types.AnthropicResponse{
			Response: response,
		},
		Success: true,
		Message: "Message processed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiResponse)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := types.APIResponse[types.HealthResponse]{
		Data: types.HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
		},
		Success: true,
		Message: "Anthropic Agent API is running",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getEndpoints(w http.ResponseWriter, r *http.Request) {
	endpoints := []types.EndpointInfo{
		{
			Path:        "/health",
			Method:      "GET",
			Description: "Health check endpoint",
			Example: map[string]interface{}{
				"data": map[string]interface{}{
					"status":    "healthy",
					"timestamp": "2025-07-04T21:14:18.055694798Z",
				},
				"success": true,
				"message": "Anthropic Agent API is running",
			},
		},
		{
			Path:        "/api/chat",
			Method:      "POST",
			Description: "Send a message to the Anthropic agent",
			Example: map[string]interface{}{
				"request": map[string]interface{}{
					"message": "Hello, how are you?",
				},
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"response": "Hello! I'm doing well, thank you for asking. I'm here to help you with any questions or tasks you might have. How can I assist you today?",
					},
					"success": true,
					"message": "Message processed successfully",
				},
			},
		},
	}

	response := types.APIResponse[types.EndpointsResponse]{
		Data: types.EndpointsResponse{
			Endpoints: endpoints,
			Total:     len(endpoints),
		},
		Success: true,
		Message: "API endpoints retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
