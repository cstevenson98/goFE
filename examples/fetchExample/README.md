# Type-Safe Fetch API Example

This example demonstrates the enhanced fetch API with type-safe HTTP requests using Go generics. It shows how frontend and backend can share the same types for seamless type safety across the full stack.

## Features

- **Type-Safe HTTP Requests**: Using Go generics for compile-time type checking
- **Shared Types**: Frontend and backend use identical type definitions
- **CRUD Operations**: Create, read, update, delete operations on users and messages
- **Real-time Updates**: Automatic UI updates when data changes
- **Error Handling**: Comprehensive error handling with user feedback

## Architecture

### Shared Types (`pkg/shared/types.go`)
Both frontend and backend use the same type definitions:

```go
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
```

### Enhanced Fetch API (`pkg/goFE/utils/http.go`)
Type-safe HTTP utilities with generics:

```go
// GetJSON performs a GET request and returns JSON data
func GetJSON[T any](url string) (*FetchResponse[T], error)

// PostJSON performs a POST request with JSON data
func PostJSON[T any, U any](url string, data U) (*FetchResponse[T], error)
```

## Usage Examples

### Frontend Type-Safe API Calls

```go
// Load users with type safety
response, err := utils.GetJSON[shared.APIResponse[shared.UsersResponse]]("http://localhost:8080/api/users")
if err != nil {
    // Handle error
    return
}

if response.OK && response.Data.Success {
    users := response.Data.Data.Users
    // Use the strongly-typed users slice
}

// Create user with type safety
request := shared.CreateUserRequest{
    Name:  "John Doe",
    Email: "john@example.com",
}

response, err := utils.PostJSON[shared.APIResponse[shared.UserResponse], shared.CreateUserRequest](
    "http://localhost:8080/api/users", 
    request,
)
```

### Backend Type-Safe Responses

```go
// Backend uses the same types
func createUser(w http.ResponseWriter, r *http.Request) {
    var request shared.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Create new user
    newUser := shared.User{
        ID:        uuid.New().String(),
        Email:     request.Email,
        Name:      request.Name,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    response := shared.APIResponse[shared.UserResponse]{
        Data: shared.UserResponse{
            User: newUser,
        },
        Success: true,
        Message: "User created successfully",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## Benefits of Type-Safe API

1. **Compile-Time Safety**: Type mismatches are caught at compile time
2. **IDE Support**: Full autocomplete and refactoring support
3. **Shared Contracts**: Frontend and backend share the same data contracts
4. **Reduced Bugs**: Eliminates runtime type errors
5. **Better Documentation**: Types serve as living documentation

## Running the Example

### 1. Start the Backend API

```bash
cd backends/api-example
go mod tidy
go run main.go
```

The API server will start on `http://localhost:8080`

### 2. Build and Run the Frontend

```bash
# Build the WASM binary
make fetch

# Serve the frontend (from the root directory)
python -m http.server 8081
```

### 3. Open the Application

Navigate to `http://localhost:8081` to see the application.

## API Endpoints

### Users
- `GET /api/users` - Get all users
- `POST /api/users` - Create a new user
- `GET /api/users/{id}` - Get a specific user
- `PUT /api/users/{id}` - Update a user
- `DELETE /api/users/{id}` - Delete a user

### Messages
- `GET /api/messages` - Get all messages (with pagination)
- `POST /api/messages` - Create a new message
- `GET /api/messages/{id}` - Get a specific message
- `DELETE /api/messages/{id}` - Delete a message

## Type Safety in Action

The example demonstrates several key benefits:

### 1. Request/Response Type Matching
```go
// Frontend sends CreateUserRequest
request := shared.CreateUserRequest{...}

// Backend receives the same type
var request shared.CreateUserRequest
json.NewDecoder(r.Body).Decode(&request)

// Backend responds with UserResponse
response := shared.APIResponse[shared.UserResponse]{...}

// Frontend receives the same type
response, err := utils.PostJSON[shared.APIResponse[shared.UserResponse], shared.CreateUserRequest](...)
```

### 2. Generic API Functions
```go
// Same function works for different types
usersResponse, err := utils.GetJSON[shared.APIResponse[shared.UsersResponse]](url)
messagesResponse, err := utils.GetJSON[shared.APIResponse[shared.MessagesResponse]](url)
```

### 3. Compile-Time Validation
If you try to use the wrong types, the compiler will catch it:

```go
// This would cause a compile error if the types don't match
response, err := utils.PostJSON[shared.User, shared.Message](url, messageData)
```

## Error Handling

The enhanced fetch API provides comprehensive error handling:

```go
response, err := utils.GetJSON[shared.APIResponse[shared.UsersResponse]](url)
if err != nil {
    // Handle network errors, parsing errors, etc.
    return
}

if !response.OK {
    // Handle HTTP errors (4xx, 5xx)
    return
}

if !response.Data.Success {
    // Handle application-level errors
    return
}
```

## Future Enhancements

1. **Request Validation**: Add validation tags to request types
2. **Caching**: Implement response caching with type safety
3. **Retry Logic**: Add automatic retry for failed requests
4. **Request Interceptors**: Add middleware for authentication, logging, etc.
5. **Response Transformers**: Add utilities for transforming responses

This example demonstrates how Go's type system can be leveraged to create a robust, type-safe API layer that works seamlessly between frontend and backend. 