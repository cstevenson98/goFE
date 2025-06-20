package shared

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// UserResponse represents a user response
type UserResponse struct {
	User User `json:"user"`
}

// UsersResponse represents a list of users response
type UsersResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

// Message represents a message in the system
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateMessageRequest represents a request to create a new message
type CreateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
	UserID  string `json:"user_id" validate:"required"`
}

// MessageResponse represents a message response
type MessageResponse struct {
	Message Message `json:"message"`
}

// MessagesResponse represents a list of messages response
type MessagesResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
}

// APIResponse represents a generic API response
type APIResponse[T any] struct {
	Data    T      `json:"data"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse[T any] struct {
	Data       []T                `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
