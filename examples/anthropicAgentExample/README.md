# Anthropic Agent Example with LilyPond

This example demonstrates how to use the goFE framework with an Anthropic AI agent for music composition using LilyPond notation. The example now supports both regular chat mode and a specialized music assistant mode.

## Features

- **Dual API Modes**: Switch between regular chat and specialized music assistant
- **AI-Powered Music Composition**: Chat with an Anthropic AI agent to help compose music
- **LilyPond Editor**: Edit LilyPond music notation in real-time
- **Document Management**: Create, save, compile, and manage LilyPond documents
- **Streaming Responses**: Get real-time streaming responses from the AI agent
- **PDF Preview**: View compiled music scores (when backend supports it)
- **Music Assistant Integration**: Specialized music composition, analysis, and suggestions

## Architecture

The example consists of:

1. **Frontend (goFE)**: A web-based interface built with the goFE framework
2. **Backend (Anthropic Agent)**: A Go server that handles AI interactions and LilyPond processing
3. **LilyPond Integration**: Music notation compilation and rendering
4. **Music Assistant**: Specialized AI assistant for music composition tasks

## Components

### Frontend Components

- `AnthropicAgentExample`: Main component that orchestrates the UI
- Template: HTML template with Tailwind CSS styling
- Event Handlers: Manage user interactions and API calls
- API Mode Toggle: Switch between chat and music assistant modes

### Backend Services

- **Anthropic Agent**: Handles AI chat interactions
- **Music Assistant**: Specialized music composition assistant
- **LilyPond Processor**: Compiles and validates LilyPond notation
- **Context Manager**: Manages document storage and retrieval
- **Prompt Engine**: Enhances prompts for better AI responses

## API Endpoints

The example communicates with the backend via these endpoints:

### Regular Chat Mode
- `POST /api/chat` - Send messages to the AI agent (includes current LilyPond content for context)
- `POST /api/chat/stream` - Set up streaming chat sessions
- `GET /api/chat/stream/{sessionId}` - Stream AI responses

### Music Assistant Mode
- `POST /api/music-assistant/chat` - Chat with music assistant (compose, analyze, modify, suggest)
- `POST /api/music-assistant/stream` - Set up streaming music assistant sessions
- `GET /api/music-assistant/stream/{sessionId}` - Stream music assistant responses
- `POST /api/music-assistant/validate-compile` - Validate and compile LilyPond content

### Document Management
- `GET /api/lilypond` - List LilyPond documents
- `POST /api/lilypond` - Create new LilyPond document
- `PUT /api/lilypond/{id}` - Update LilyPond document
- `POST /api/lilypond/{id}/compile` - Compile LilyPond to PDF
- `DELETE /api/lilypond/{id}` - Delete LilyPond document

## Usage

1. **Start the Backend**: Run the anthropic-agent backend server
2. **Start the Frontend**: Run this example with `go run main.go`
3. **Choose API Mode**: Use the toggle button to switch between:
   - **Chat Mode**: General conversation with the AI
   - **Music Assistant Mode**: Specialized music composition assistance
4. **Compose Music**: Use the chat interface to ask for music composition help
5. **Edit Notation**: Modify the LilyPond code in the editor
6. **Compile**: Click "Compile" to generate PDF output

## Music Assistant Features

The music assistant mode provides specialized capabilities:

- **Composition**: Create new musical pieces with specific styles and parameters
- **Analysis**: Analyze existing pieces for harmonic, melodic, and rhythmic content
- **Modification**: Modify pieces with specific instructions (add dynamics, change key, etc.)
- **Suggestions**: Get improvement suggestions for musical pieces
- **Context Awareness**: Understand musical context and provide relevant advice

## Example Prompts

### Music Assistant Mode
- "Create a simple C major scale in 4/4 time"
- "Analyze this musical piece for harmonic structure"
- "Add dynamics and articulation to this melody"
- "Suggest improvements for this composition"
- "Modify this piece to be in 3/4 time"

### Chat Mode (Context-Aware)
- "Explain the difference between major and minor scales"
- "What is LilyPond notation?"
- "How do I write a chord progression?"
- "Add a bass line to this piece" (uses current content)
- "Change the key of this piece to G major" (uses current content)
- "What's wrong with this LilyPond code?" (uses current content)
- "Make this piece more dramatic" (uses current content)

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

## API Response Types

The example handles different response types:

- **AnthropicResponse**: Basic chat responses
- **MusicAssistantResponse**: Enhanced responses with LilyPond content, compilation results, and analysis
- **CompileResult**: LilyPond compilation results with errors and warnings
- **StreamSetupResponse**: Streaming session setup responses 