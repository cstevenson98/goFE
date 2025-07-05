# Anthropic Agent API

A simple REST API that wraps the Anthropic Claude API, providing a clean interface for sending messages and receiving responses.

## Features

- REST API endpoint for sending messages to Claude
- Health check endpoint
- API documentation endpoint
- CORS support
- Error handling with consistent response format
- Docker support

## Setup

### Prerequisites

- Go 1.21 or later
- Anthropic API key

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

The server will start on port 8081.

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

Send a message to the Anthropic agent.

**Request:**
```json
{
  "message": "Hello, how are you?"
}
```

**Response:**
```json
{
  "data": {
    "response": "Hello! I'm doing well, thank you for asking..."
  },
  "success": true,
  "message": "Message processed successfully"
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2025-07-04T21:14:18.055694798Z"
  },
  "success": true,
  "message": "Anthropic Agent API is running"
}
```

### GET /endpoints

Get API documentation.

**Response:**
```json
{
  "data": {
    "endpoints": [
      {
        "path": "/health",
        "method": "GET",
        "description": "Health check endpoint"
      },
      {
        "path": "/api/chat",
        "method": "POST",
        "description": "Send a message to the Anthropic agent"
      }
    ],
    "total": 2
  },
  "success": true,
  "message": "API endpoints retrieved successfully"
}
```

## Architecture

The service follows a clean architecture pattern:

- **Interface**: `AnthropicAgent` defines the contract for interacting with Anthropic's API
- **Implementation**: `anthropicAgent` wraps the official Anthropic Go SDK
- **API Layer**: REST endpoints that use the agent interface
- **Response Format**: Consistent API response structure matching the existing api-example pattern

This design allows for easy testing and potential future changes to the underlying AI provider. 