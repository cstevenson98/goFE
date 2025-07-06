# Music Composition Assistant API

A specialized REST API that provides a music composition assistant powered by Anthropic's Claude API. The assistant helps users create musical scores using LilyPond notation, with support for document management, LilyPond processing, and intelligent music composition guidance.

## Features

- **Music Composition Assistant**: AI-powered assistant specialized in music theory and LilyPond notation
- **Document Management**: Create, retrieve, update, and delete music documents
- **LilyPond Processing**: Handle LilyPond documents with source code retrieval
- **Context Management**: Maintain document context and history
- **Prompt Engineering**: Intelligent prompt construction for music composition
- **REST API**: Clean interface for all operations
- **Real-time Streaming**: Server-Sent Events (SSE) for streaming responses
- **Health Check**: API status monitoring
- **API Documentation**: Built-in endpoint documentation
- **CORS Support**: Cross-origin resource sharing
- **Error Handling**: Consistent error response format
- **Docker Support**: Containerized deployment

## Setup

### Prerequisites

- Go 1.21 or later
- Anthropic API key
- LilyPond (for PDF compilation):
  - **Ubuntu/Debian**: `sudo apt-get install lilypond`
  - **macOS**: `brew install lilypond`
  - **Windows**: Download from https://lilypond.org/download.html
  
  The system will use the `lilypond` command to compile music notation to PDF.

### Environment Variables

Set the following environment variable:

```bash
export ANTHROPIC_API_KEY="your-anthropic-api-key-here"
```

### Running Locally

1. Navigate to the anthropic-agent directory:
```bash
cd backends/anthropic-agent
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the server:
```bash
go run .
```

The server will start on port 8081 as the Music Composition Assistant API.

### Running with Docker

1. Build the Docker image:
```bash
docker build -t anthropic-agent .
```

2. Run the container:
```bash
docker run -p 8081:8081 -e ANTHROPIC_API_KEY="your-api-key" anthropic-agent
```

## API Endpoints

### POST /api/chat

Send a message to the music composition assistant and receive a complete response.

**Request:**
```json
{
  "message": "Create a simple C major scale in LilyPond notation"
}
```

**Response:**
```json
{
  "data": {
    "response": "I'll help you create a C major scale in LilyPond notation. Here's a complete example:\\n\\n\\version \"2.24.0\"\\n\\score {\\n  \\new Staff {\\n    c d e f g a b c\\n  }\\n  \\layout {}\\n}\\n\\nThis creates a C major scale with quarter notes. The notes c, d, e, f, g, a, b represent the notes of the C major scale."
  },
  "success": true,
  "message": "Message processed successfully"
}
```

**Error Response:**
```json
{
  "data": {},
  "success": false,
  "error": "Failed to send message to Claude: API key not found"
}
```

### POST /api/chat/stream

Set up a streaming session for real-time message responses via Server-Sent Events.

**Request:**
```json
{
  "message": "What is a quaternion?"
}
```

**Response:**
```json
{
  "data": {
    "sessionId": "session_1703123456789012345"
  },
  "success": true,
  "message": "Stream session created"
}
```

**Error Response:**
```json
{
  "data": {},
  "success": false,
  "error": "Message is required"
}
```

### GET /api/chat/stream/{sessionId}

Connect to a Server-Sent Events stream to receive real-time streaming responses.

**URL:** `/api/chat/stream/session_1703123456789012345`

**Headers:**
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

**Event Types:**

1. **Message Events** - Streaming text chunks:
```
event: message
data: A quaternion is a mathematical object that extends the concept of complex numbers to four dimensions.
```

2. **Complete Event** - Stream finished successfully:
```
event: complete
data: {}
```

3. **Error Event** - Error occurred during streaming:
```
event: error
data: Failed to send message to Claude: API key not found
```

**JavaScript Client Example:**
```javascript
// First, set up the stream session
const setupResponse = await fetch('/api/chat/stream', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ message: 'What is a quaternion?' })
});
const { data: { sessionId } } = await setupResponse.json();

// Then connect to the SSE stream
const eventSource = new EventSource(`/api/chat/stream/${sessionId}`);

eventSource.addEventListener('message', (event) => {
  console.log('Received chunk:', event.data);
  // Append to UI or process the streaming text
});

eventSource.addEventListener('complete', (event) => {
  console.log('Stream completed');
  eventSource.close();
});

eventSource.addEventListener('error', (event) => {
  console.error('Stream error:', event.data);
  eventSource.close();
});
```

### POST /api/documents

Create a new document for storing music content.

**Request:**
```json
{
  "title": "My Music Score",
  "content": "\\version \"2.24.0\"\\score {\\new Staff {c d e f g a b c}\\layout {}}"
}
```

**Response:**
```json
{
  "data": {
    "id": "doc_1234567890",
    "title": "My Music Score",
    "content": "\\version \"2.24.0\"\\score {\\new Staff {c d e f g a b c}\\layout {}}",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "success": true,
  "message": "Document created successfully"
}
```

### GET /api/documents

List all documents with pagination support.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)

**Response:**
```json
{
  "data": {
    "documents": [
      {
        "id": "doc_1234567890",
        "title": "My Music Score",
        "content": "\\documentclass{article}\\begin{document}Music content here\\end{document}",
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "totalPages": 1
    }
  },
  "success": true,
  "message": "Documents retrieved successfully"
}
```

### GET /api/documents/{id}

Retrieve a specific document by ID.

**Response:**
```json
{
  "data": {
    "id": "doc_1234567890",
    "title": "My Music Score",
    "content": "\\documentclass{article}\\begin{document}Music content here\\end{document}",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "success": true,
  "message": "Document retrieved successfully"
}
```

### PUT /api/documents/{id}

Update an existing document.

**Request:**
```json
{
  "title": "Updated Music Score",
  "content": "\\documentclass{article}\\begin{document}Updated music content\\end{document}"
}
```

### DELETE /api/documents/{id}

Delete a document.

**Response:**
```json
{
  "data": {
    "id": "doc_1234567890"
  },
  "success": true,
  "message": "Document deleted successfully"
}
```

### POST /api/lilypond

Create a new LilyPond document for music notation.

**Request:**
```json
{
  "title": "Music Score",
  "content": "\\version \"2.24.0\"\\score {\\new Staff {c d e f g a b c}\\layout {}}"
}
```

**Response:**
```json
{
  "data": {
    "id": "lilypond_1234567890",
    "title": "Music Score",
    "content": "\\version \"2.24.0\"\\score {\\new Staff {c d e f g a b c}\\layout {}}",
    "status": "draft",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "success": true,
  "message": "LilyPond document created successfully"
}
```

### GET /api/lilypond/{id}/source

Get the LilyPond source code for a document.

**Response:**
```json
{
  "data": {
    "id": "lilypond_1234567890",
    "source": "\\version \"2.24.0\"\\score {\\new Staff {c d e f g a b c}\\layout {}}"
  },
  "success": true,
  "message": "LilyPond source code retrieved successfully"
}
```

### GET /api/lilypond/{id}/file-path

Get the file system path where the LilyPond document is stored.

**Response:**
```json
{
  "data": {
    "id": "lilypond_1234567890",
    "file_path": "/tmp/music_composition_assistant/lilypond/lilypond_1234567890.ly"
  },
  "success": true,
  "message": "LilyPond file path retrieved successfully"
}
```

### GET /api/lilypond/temp-dir

Get the temporary directory where LilyPond files are stored.

**Response:**
```json
{
  "data": {
    "temp_dir": "/tmp/music_composition_assistant/lilypond"
  },
  "success": true,
  "message": "LilyPond temporary directory retrieved successfully"
}
```

### POST /api/lilypond/{id}/compile

Compile a LilyPond document to PDF.

**Response:**
```json
{
  "data": {
    "id": "lilypond_1234567890",
    "status": "compiled",
    "pdf_path": "/tmp/music_composition_assistant/lilypond/lilypond_1234567890.pdf"
  },
  "success": true,
  "message": "LilyPond document compiled successfully"
}
```

**Error Response:**
```json
{
  "data": {},
  "success": false,
  "error": "LilyPond compilation failed: lilypond: command not found"
}
```

### GET /api/lilypond/{id}/pdf

Download the compiled PDF file.

**Response:** Binary PDF data with `Content-Type: application/pdf`

### GET /api/lilypond/{id}/pdf-path

Get the file system path where the compiled PDF is stored.

**Response:**
```json
{
  "data": {
    "id": "lilypond_1234567890",
    "pdf_path": "/tmp/music_composition_assistant/lilypond/lilypond_1234567890.pdf"
  },
  "success": true,
  "message": "LilyPond PDF path retrieved successfully"
}
```

### DELETE /api/lilypond/{id}

Delete a LilyPond document and its associated files (both .ly and .pdf).

**Response:**
```json
{
  "data": {
    "id": "lilypond_1234567890"
  },
  "success": true,
  "message": "LilyPond document deleted successfully"
}
```

### GET /health

Health check endpoint to verify the API is running.

**Response:**
```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2025-07-04T21:14:18.055694798Z"
  },
  "success": true,
  "message": "Music Composition Assistant API is running"
}
```

### GET /endpoints

Get comprehensive API documentation with examples.

**Response:**
```json
{
  "data": {
    "endpoints": [
      {
        "path": "/health",
        "method": "GET",
        "description": "Health check endpoint",
        "example": {
          "data": {
            "status": "healthy",
            "timestamp": "2025-07-04T21:14:18.055694798Z"
          },
          "success": true,
          "message": "Anthropic Agent API is running"
        }
      },
      {
        "path": "/api/chat",
        "method": "POST",
        "description": "Send a message to the Anthropic agent",
        "example": {
          "request": {
            "message": "Hello, how are you?"
          },
          "response": {
            "data": {
              "response": "Hello! I'm doing well, thank you for asking..."
            },
            "success": true,
            "message": "Message processed successfully"
          }
        }
      },
      {
        "path": "/api/chat/stream",
        "method": "POST",
        "description": "Set up a streaming session for real-time responses",
        "example": {
          "request": {
            "message": "What is a quaternion?"
          },
          "response": {
            "data": {
              "sessionId": "session_1703123456789012345"
            },
            "success": true,
            "message": "Stream session created"
          }
        }
      },
      {
        "path": "/api/chat/stream/{sessionId}",
        "method": "GET",
        "description": "Connect to Server-Sent Events stream for real-time responses"
      }
    ],
    "total": 4
  },
  "success": true,
  "message": "API endpoints retrieved successfully"
}
```

## Request/Response Types

### AnthropicRequest
```go
type AnthropicRequest struct {
    Message string `json:"message" validate:"required,min=1,max=10000"`
}
```

### AnthropicResponse
```go
type AnthropicResponse struct {
    Response string `json:"response"`
}
```

### StreamSetupResponse
```go
type StreamSetupResponse struct {
    SessionId string `json:"sessionId"`
}
```

### APIResponse
```go
type APIResponse[T any] struct {
    Data    T      `json:"data"`
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

## Architecture

The service follows a clean architecture pattern with the following components:

### Core Components
- **Context Management**: `ContextManager` handles document storage and retrieval
- **LilyPond Processing**: `LilyPondProcessor` manages LilyPond documents with file system integration
- **Prompt Engineering**: `PromptEngine` constructs intelligent prompts for music composition
- **Agent Interface**: `AnthropicAgent` defines the contract for interacting with Anthropic's API
- **Agent Implementation**: `anthropicAgent` wraps the official Anthropic Go SDK

### File System Integration
- **Temporary Directory**: LilyPond files are stored in a temporary directory (`/tmp/music_composition_assistant/lilypond/` on Linux)
- **Unique File Names**: Each document gets a unique ID-based filename (e.g., `lilypond_1234567890.ly`)
- **PDF Generation**: Compiled PDFs are stored alongside LilyPond files with matching names (e.g., `lilypond_1234567890.pdf`)
- **File Operations**: Create, read, update, and delete operations are synchronized with the file system
- **Fallback Handling**: If the system temp directory is unavailable, falls back to local directories

### LilyPond Compilation
- **LilyPond Compiler**: Uses the `lilypond` command to compile music notation to PDF
- **Compilation Process**: Changes to document directory, runs LilyPond compiler, and verifies PDF creation
- **Status Tracking**: Documents track compilation status (draft, compiled, error)
- **Error Handling**: Comprehensive error reporting for compilation failures

### API Layer
- **REST Endpoints**: Clean REST API for all operations
- **Streaming**: `streamingReader` implements `io.Reader` for real-time text streaming
- **SSE Support**: Server-Sent Events for real-time streaming responses
- **Response Format**: Consistent API response structure across all endpoints

### Data Flow
1. **Document Creation**: Content is validated, stored in memory, and written to file system
2. **Document Retrieval**: Content is read from file system to ensure consistency
3. **Document Updates**: Both memory and file system are updated atomically
4. **Document Deletion**: Both memory and file system are cleaned up

## Testing

Use the provided test script to verify all endpoints:

```bash
chmod +x test_music_assistant.sh
./test_music_assistant.sh
```

The test script will verify:
- Health check endpoint
- Document creation and retrieval
- LaTeX document operations
- Music composition assistant chat
- API documentation

## Error Handling

All endpoints return consistent error responses with:
- `success: false` for errors
- `error` field containing the error message
- Appropriate HTTP status codes

## CORS Support

The API includes CORS middleware that allows:
- All origins (`*`)
- Common HTTP methods (GET, POST, PUT, DELETE, OPTIONS)
- Standard headers (Content-Type, Authorization)

## Session Management

Streaming sessions are managed in-memory with automatic cleanup:
- Sessions are created with unique IDs
- Sessions are automatically cleaned up when the stream completes or errors
- Session storage uses thread-safe operations with mutex locks

This design allows for easy testing and potential future changes to the underlying AI provider while providing both traditional and real-time streaming capabilities. 