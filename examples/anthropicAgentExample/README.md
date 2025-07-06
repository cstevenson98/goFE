# Anthropic Agent Example with LilyPond

This example demonstrates how to use the goFE framework with an Anthropic AI agent for music composition using LilyPond notation.

## Features

- **AI-Powered Music Composition**: Chat with an Anthropic AI agent to help compose music
- **LilyPond Editor**: Edit LilyPond music notation in real-time
- **Document Management**: Create, save, compile, and manage LilyPond documents
- **Streaming Responses**: Get real-time streaming responses from the AI agent
- **PDF Preview**: View compiled music scores (when backend supports it)

## Architecture

The example consists of:

1. **Frontend (goFE)**: A web-based interface built with the goFE framework
2. **Backend (Anthropic Agent)**: A Go server that handles AI interactions and LilyPond processing
3. **LilyPond Integration**: Music notation compilation and rendering

## Components

### Frontend Components

- `AnthropicAgentExample`: Main component that orchestrates the UI
- Template: HTML template with Tailwind CSS styling
- Event Handlers: Manage user interactions and API calls

### Backend Services

- **Anthropic Agent**: Handles AI chat interactions
- **LilyPond Processor**: Compiles and validates LilyPond notation
- **Context Manager**: Manages document storage and retrieval
- **Prompt Engine**: Enhances prompts for better AI responses

## API Endpoints

The example communicates with the backend via these endpoints:

- `POST /api/chat` - Send messages to the AI agent
- `POST /api/chat/stream` - Set up streaming chat sessions
- `GET /api/chat/stream/{sessionId}` - Stream AI responses
- `GET /api/lilypond` - List LilyPond documents
- `POST /api/lilypond` - Create new LilyPond document
- `PUT /api/lilypond/{id}` - Update LilyPond document
- `POST /api/lilypond/{id}/compile` - Compile LilyPond to PDF
- `DELETE /api/lilypond/{id}` - Delete LilyPond document

## Usage

1. **Start the Backend**: Run the anthropic-agent backend server
2. **Start the Frontend**: Run this example with `go run main.go`
3. **Compose Music**: Use the chat interface to ask for music composition help
4. **Edit Notation**: Modify the LilyPond code in the editor
5. **Compile**: Click "Compile" to generate PDF output

## Example LilyPond Code

```lilypond
\version "2.24.0"

\header {
  title = "Simple Melody"
  composer = "Music Assistant"
}

\score {
  \new Staff {
    \clef treble
    \time 4/4
    \key c \major
    
    c'4 d'4 e'4 f'4 |
    g'4 a'4 b'4 c''4 |
  }
}
```

## Dependencies

- goFE framework
- Anthropic AI API
- LilyPond (for music notation compilation)
- Tailwind CSS (for styling)

## Configuration

Make sure the backend server is running on `localhost:8081` before starting this example. 