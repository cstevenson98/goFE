package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/agent"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/core"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/core/lilypond"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/core/prompt"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/handlers"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/types"
	"github.com/gorilla/mux"
)

// Global instances
var agentInstance agent.AnthropicAgent
var lilypondProcessor *lilypond.LilyPondProcessor
var promptEngine *prompt.PromptEngine
var musicAssistant *core.MusicAssistant

// Stream sessions storage
var streamSessions = make(map[string]string)
var streamMutex sync.RWMutex

func main() {
	// Initialize the Anthropic agent
	agentInstance = agent.NewAnthropicAgent()
	if err := agentInstance.Instantiate(); err != nil {
		log.Fatalf("Failed to initialize Anthropic agent: %v", err)
	}

	// Initialize LilyPond processor
	lilypondProcessor = lilypond.NewLilyPondProcessor()

	// Initialize prompt engine with LilyPond version
	promptEngine = prompt.NewPromptEngineWithVersion(lilypondProcessor.GetVersion())

	// Initialize music assistant
	musicAssistant = core.NewMusicAssistant(lilypondProcessor, promptEngine, agentInstance)

	// Initialize handlers
	lilypondHandler := handlers.NewLilyPondHandler(lilypondProcessor)
	musicAssistantHandler := handlers.NewMusicAssistantHandler(musicAssistant)

	r := mux.NewRouter()

	// Apply CORS middleware first
	r.Use(corsMiddleware)

	// Anthropic agent routes
	r.HandleFunc("/api/chat", handleChat).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/chat/stream", handleStreamSetup).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/chat/stream/{sessionId}", handleStreamEvents).Methods("GET")

	// LilyPond document routes
	r.HandleFunc("/api/lilypond", lilypondHandler.ListLilyPondDocuments).Methods("GET")
	r.HandleFunc("/api/lilypond", lilypondHandler.CreateLilyPondDocument).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/lilypond/{id}", lilypondHandler.GetLilyPondDocument).Methods("GET")
	r.HandleFunc("/api/lilypond/{id}", lilypondHandler.UpdateLilyPondDocument).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/lilypond/{id}", lilypondHandler.DeleteLilyPondDocument).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/lilypond/{id}/compile", lilypondHandler.CompileLilyPondDocument).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/lilypond/{id}/pdf", lilypondHandler.GetLilyPondPDF).Methods("GET")

	// Music Assistant routes
	r.HandleFunc("/api/music-assistant/chat", musicAssistantHandler.HandleChat).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/music-assistant/stream", musicAssistantHandler.HandleStreamSetup).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/music-assistant/stream/{sessionId}", musicAssistantHandler.HandleStreamEvents).Methods("GET")
	r.HandleFunc("/api/music-assistant/validate-compile", musicAssistantHandler.HandleValidateAndCompile).Methods("POST", "OPTIONS")

	// Health check endpoint
	r.HandleFunc("/health", healthCheck).Methods("GET")

	// Endpoints documentation
	r.HandleFunc("/endpoints", getEndpoints).Methods("GET")

	fmt.Println("Music Composition Assistant API server starting on :8081")
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

	// Build enhanced prompt using the prompt engine
	context := map[string]interface{}{
		"context": "music_composition",
	}
	if request.Content != "" {
		context["lilypond_content"] = request.Content
	}
	enhancedPrompt := promptEngine.BuildPrompt(request.Message, context)

	// Send message to Anthropic agent with enhanced prompt
	response, err := agentInstance.SendMessage(enhancedPrompt)
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

func handleStreamSetup(w http.ResponseWriter, r *http.Request) {
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

	// Generate session ID
	sessionId := fmt.Sprintf("session_%d", time.Now().UnixNano())

	// Store message in session
	streamMutex.Lock()
	streamSessions[sessionId] = request.Message
	streamMutex.Unlock()

	// Return session ID
	response := types.APIResponse[types.StreamSetupResponse]{
		Data: types.StreamSetupResponse{
			SessionId: sessionId,
		},
		Success: true,
		Message: "Stream session created",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleStreamEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionId := vars["sessionId"]

	// Get message from session
	streamMutex.RLock()
	message, exists := streamSessions[sessionId]
	streamMutex.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create a channel to signal when the client disconnects
	notify := r.Context().Done()

	// Get the reader from the agent
	reader, err := agentInstance.SendMessageStream(message)
	if err != nil {
		// Send error event
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		w.(http.Flusher).Flush()
		return
	}

	// Clean up session
	defer func() {
		streamMutex.Lock()
		delete(streamSessions, sessionId)
		streamMutex.Unlock()
	}()

	// Buffer to read chunks
	buffer := make([]byte, 1024)

	// Stream the response
	for {
		select {
		case <-notify:
			// Client disconnected
			return
		default:
			n, err := reader.Read(buffer)
			if n > 0 {
				// Send the chunk as an event
				chunk := string(buffer[:n])
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", chunk)
				w.(http.Flusher).Flush()
			}
			if err != nil {
				if err.Error() == "EOF" {
					// Send completion event
					fmt.Fprintf(w, "event: complete\ndata: {}\n\n")
					w.(http.Flusher).Flush()
				} else {
					// Send error event
					fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
					w.(http.Flusher).Flush()
				}
				return
			}
		}
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := types.APIResponse[types.HealthResponse]{
		Data: types.HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
		},
		Success: true,
		Message: "Music Composition Assistant API is running",
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
				"message": "Music Composition Assistant API is running",
			},
		},
		{
			Path:        "/api/chat",
			Method:      "POST",
			Description: "Send a message to the music composition assistant (can include current LilyPond content)",
			Example: map[string]interface{}{
				"request": map[string]interface{}{
					"message": "Create a simple C major scale in LilyPond notation",
					"content": "\\version \"2.22.1\"\\score{\\new Staff{c d e f}}\\layout{}",
				},
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"response": "I'll help you create a C major scale in LilyPond notation...",
					},
					"success": true,
					"message": "Message processed successfully",
				},
			},
		},
		{
			Path:        "/api/music-assistant/chat",
			Method:      "POST",
			Description: "Chat with the music assistant (compose, analyze, modify, suggest)",
			Example: map[string]interface{}{
				"request": map[string]interface{}{
					"message": "Create a simple melody in C major",
					"context": map[string]string{
						"style": "classical",
						"tempo": "moderate",
					},
				},
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"response":   "I'll create a simple C major melody for you...",
						"newContent": "\\version \"2.24.0\"\\score{...}",
						"message":    "LilyPond content generated successfully",
					},
					"success": true,
					"message": "Response processed successfully",
				},
			},
		},
		{
			Path:        "/api/music-assistant/stream",
			Method:      "POST",
			Description: "Set up streaming session with music assistant",
			Example: map[string]interface{}{
				"request": map[string]interface{}{
					"message": "Analyze this musical piece",
					"context": map[string]string{
						"focus": "harmonic analysis",
					},
				},
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"sessionId": "music_session_1234567890",
					},
					"success": true,
					"message": "Stream session created",
				},
			},
		},
		{
			Path:        "/api/music-assistant/validate-compile",
			Method:      "POST",
			Description: "Validate and compile LilyPond content",
			Example: map[string]interface{}{
				"request": map[string]interface{}{
					"content": "\\version \"2.24.0\"\\score{...}",
				},
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"success":       true,
						"newContent":    "\\version \"2.24.0\"\\score{...}",
						"compileResult": map[string]interface{}{},
						"message":       "Content validated and compiled successfully",
					},
					"success": true,
					"message": "Validation and compilation completed successfully",
				},
			},
		},
		{
			Path:        "/api/lilypond",
			Method:      "POST",
			Description: "Create a new LilyPond document",
			Example: map[string]interface{}{
				"request": map[string]interface{}{
					"title":   "My Music Score",
					"content": "\\version \"2.24.0\"\\score{...}",
				},
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"id":         "lilypond_1234567890",
						"title":      "My Music Score",
						"content":    "\\version \"2.24.0\"\\score{...}",
						"status":     "draft",
						"created_at": "2025-01-01T00:00:00Z",
						"updated_at": "2025-01-01T00:00:00Z",
					},
					"success": true,
					"message": "LilyPond document created successfully",
				},
			},
		},
		{
			Path:        "/api/lilypond",
			Method:      "GET",
			Description: "List all LilyPond documents",
			Example: map[string]interface{}{
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"documents": []map[string]interface{}{},
						"pagination": map[string]interface{}{
							"page":       1,
							"limit":      10,
							"total":      0,
							"totalPages": 0,
						},
					},
					"success": true,
					"message": "LilyPond documents retrieved successfully",
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
