package types

import "time"

// AnthropicRequest represents a request to the Anthropic agent
type AnthropicRequest struct {
	Message string `json:"message" validate:"required,min=1,max=10000"`
	Content string `json:"content,omitempty"` // Current LilyPond content to analyze/modify
}

// AnthropicResponse represents a response from the Anthropic agent
type AnthropicResponse struct {
	Response string `json:"response"`
}

// APIResponse represents a generic API response
type APIResponse[T any] struct {
	Data    T      `json:"data"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// EndpointInfo represents API endpoint information
type EndpointInfo struct {
	Path        string      `json:"path"`
	Method      string      `json:"method"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
}

// EndpointsResponse represents the endpoints documentation response
type EndpointsResponse struct {
	Endpoints []EndpointInfo `json:"endpoints"`
	Total     int            `json:"total"`
}

// StreamSetupResponse represents the response when setting up a streaming session
type StreamSetupResponse struct {
	SessionId string `json:"sessionId"`
}
