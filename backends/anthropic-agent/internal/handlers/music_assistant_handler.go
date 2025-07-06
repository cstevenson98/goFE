package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/core"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/types"
)

// MusicAssistantHandler handles music assistant operations
type MusicAssistantHandler struct {
	musicAssistant *core.MusicAssistant
}

// NewMusicAssistantHandler creates a new music assistant handler
func NewMusicAssistantHandler(ma *core.MusicAssistant) *MusicAssistantHandler {
	return &MusicAssistantHandler{
		musicAssistant: ma,
	}
}

// HandleChat handles music assistant chat requests
func (mah *MusicAssistantHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request core.MusicAssistantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		mah.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if request.Message == "" {
		mah.respondWithError(w, http.StatusBadRequest, "Message is required")
		return
	}

	// Set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Process the request
	response, err := mah.musicAssistant.SendMessage(ctx, &request)
	if err != nil {
		mah.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Processing failed: %v", err))
		return
	}

	// Create API response
	apiResponse := types.APIResponse[core.MusicAssistantResponse]{
		Data:    *response,
		Success: response.Error == "",
		Message: response.Message,
		Error:   response.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	if response.Error != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(apiResponse)
}

// HandleStreamSetup handles stream setup requests
func (mah *MusicAssistantHandler) HandleStreamSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request core.MusicAssistantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		mah.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if request.Message == "" {
		mah.respondWithError(w, http.StatusBadRequest, "Message is required")
		return
	}

	// Generate session ID
	sessionId := fmt.Sprintf("music_session_%d", time.Now().UnixNano())

	// Store request in session (in a real implementation, you'd use a proper session store)
	// For now, we'll just return the session ID
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

// HandleStreamEvents handles streaming events
func (mah *MusicAssistantHandler) HandleStreamEvents(w http.ResponseWriter, r *http.Request) {
	// This would be implemented similar to the existing stream events handler
	// but using the music assistant instead of the regular agent
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// For now, just send a placeholder response
	fmt.Fprintf(w, "event: message\ndata: Music assistant streaming not yet implemented\n\n")
	w.(http.Flusher).Flush()
}

// HandleValidateAndCompile handles validation and compilation requests
func (mah *MusicAssistantHandler) HandleValidateAndCompile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request struct {
		Content    string `json:"content" validate:"required"`
		DocumentID string `json:"documentId,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		mah.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.Content == "" {
		mah.respondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Validate and compile
	response, err := mah.musicAssistant.ValidateAndCompile(request.Content, request.DocumentID)
	if err != nil {
		mah.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Validation and compilation failed: %v", err))
		return
	}

	// Create API response
	apiResponse := types.APIResponse[core.MusicAssistantResponse]{
		Data:    *response,
		Success: response.Error == "",
		Message: response.Message,
		Error:   response.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	if response.Error != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(apiResponse)
}

// respondWithError sends an error response
func (mah *MusicAssistantHandler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	log.Printf("Music Assistant Error: %s", message)

	response := types.APIResponse[map[string]interface{}]{
		Data:    map[string]interface{}{},
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
