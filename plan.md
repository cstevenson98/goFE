# Detailed Plan: Anthropic Music Composer Assistant

## Overview
A code editor helper for musical notation writing using text, with a three-panel interface: prompt chat, LilyPond document editor with diff visualization, and rendered PDF preview.

## Architecture

### Backend (./backends/music-composer)
**Current State**: Basic anthropic-agent structure exists
**New Requirements**: Music-specific context management, LilyPond document handling, diff generation

### Frontend (./examples/musicComposer)
**Current State**: Basic anthropicAgentExample exists
**New Requirements**: Three-panel layout, LilyPond editor, diff visualization, PDF preview

## Detailed Implementation Plan

### Phase 1: Backend Enhancement

#### 1.1 Context Management System
```go
// internal/context/manager.go
type DocumentContext struct {
    CurrentLilyPond string
    DocumentHistory []DocumentVersion
    MusicMetadata   MusicMetadata
    ConversationHistory []ConversationTurn
}

type MusicMetadata struct {
    Title       string
    Composer    string
    Key         string
    TimeSignature string
    Tempo       string
    Instruments []string
}

type DocumentVersion struct {
    ID          string
    Timestamp   time.Time
    LilyPond    string
    Description string
    Diff        *DiffResult
}

type ConversationTurn struct {
    UserPrompt  string
    AssistantResponse string
    AppliedChanges []Change
    Timestamp   time.Time
}
```

#### 1.2 Prompt Engineering System
See [detailed prompt engineering system documentation](docs/prompt-engineering-system.md) for comprehensive implementation details.

```go
// internal/prompts/engine.go
type PromptEngine struct {
    baseSystemPrompt string
    contextWindow    int
    maxTokens        int
}

func (pe *PromptEngine) BuildMusicCompositionPrompt(
    userPrompt string, 
    context *DocumentContext,
    includeDiff bool,
) string {
    // Construct context-aware prompt with:
    // 1. System instructions for music notation
    // 2. Current LaTeX document state
    // 3. Conversation history (truncated)
    // 4. User's specific request
    // 5. Instructions for diff format
}
```

#### 1.3 Diff Generation System
See [detailed diff generation system documentation](docs/diff-generation-system.md) for comprehensive implementation details.

```go
// internal/diff/generator.go
type DiffResult struct {
    Before     string
    After      string
    Changes    []Change
    HunkRanges []HunkRange
}

type Change struct {
    Type      ChangeType // INSERT, DELETE, MODIFY
    LineNum   int
    OldText   string
    NewText   string
    Context   string
}

func GenerateDiff(before, after string) *DiffResult {
    // Use diff algorithm to generate structured diff
    // Return before/after views and change metadata
}
```

#### 1.4 LilyPond Processing System
See [detailed LilyPond processing system documentation](docs/lilypond-processing-system.md) for comprehensive implementation details.

```go
// internal/lilypond/processor.go
type LilyPondProcessor struct {
    tempDir    string
    lilypond   string
}

func (lp *LilyPondProcessor) CompileToPDF(lilypond string) ([]byte, error) {
    // 1. Write LilyPond to temp file
    // 2. Run lilypond compilation
    // 3. Return PDF bytes or error details
}

func (lp *LilyPondProcessor) ValidateSyntax(lilypond string) *ValidationResult {
    // Parse LilyPond and return syntax errors/warnings
}
```

#### 1.5 REST API Server Interface
See [detailed REST API server documentation](docs/rest-api-server.md) for comprehensive implementation details.

The server provides a collection of handlers for core business logic:
- **PromptHandler** - Prompt generation and response parsing
- **DiffHandler** - Diff generation and rendering
- **LilyPondHandler** - LilyPond compilation and validation
- **ContextHandler** - Document context management

```go
// Handler interface
type Handler interface {
    RegisterRoutes(router *mux.Router)
    GetName() string
    GetVersion() string
    HealthCheck() error
}
```

### Phase 2: Frontend Implementation

#### 2.1 Component Structure
```
examples/musicComposer/
├── main.go
├── components/
│   ├── musicComposer/
│   │   ├── musicComposer.go      // Main container
│   │   └── musicComposer.qtpl    // Three-panel layout
│   ├── chatPanel/
│   │   ├── chatPanel.go          // Prompt chat interface
│   │   └── chatPanel.qtpl        // Chat UI
│   ├── editorPanel/
│   │   ├── editorPanel.go        // LilyPond editor with diff
│   │   └── editorPanel.qtpl      // Editor UI
│   └── previewPanel/
│       ├── previewPanel.go       // PDF preview
│       └── previewPanel.qtpl     // Preview UI
```

#### 2.2 Main Container Component
```go
// components/musicComposer/musicComposer.go
type MusicComposer struct {
    id            uuid.UUID
    chatPanel     *ChatPanel
    editorPanel   *EditorPanel
    previewPanel  *PreviewPanel
    
    // State
    currentDocument *DocumentState
    isCompiling     bool
    lastCompileTime time.Time
}

type DocumentState struct {
    LilyPondContent string
    CurrentDiff     *DiffState
    PDFUrl          string
    CompileErrors   []string
}
```

#### 2.3 Chat Panel Component
```go
// components/chatPanel/chatPanel.go
type ChatPanel struct {
    id              uuid.UUID
    messageInputID  uuid.UUID
    sendButtonID    uuid.UUID
    chatHistoryID   uuid.UUID
    
    // State
    messages        []ChatMessage
    isStreaming     bool
    currentStream   *utils.EventSource
}

type ChatMessage struct {
    ID        string
    Type      MessageType // USER, ASSISTANT
    Content   string
    Timestamp time.Time
    Applied   bool
}
```

#### 2.4 Editor Panel Component
```go
// components/editorPanel/editorPanel.go
type EditorPanel struct {
    id              uuid.UUID
    editorID        uuid.UUID
    diffViewID      uuid.UUID
    
    // State
    lilypondContent string
    diffState       *DiffState
    isEditing       bool
}

type DiffState struct {
    BeforeLines     []string
    AfterLines      []string
    ChangeMap       map[int]ChangeType
    HunkRanges      []HunkRange
}
```

#### 2.5 Preview Panel Component
```go
// components/previewPanel/previewPanel.go
type PreviewPanel struct {
    id              uuid.UUID
    pdfViewerID     uuid.UUID
    errorDisplayID  uuid.UUID
    
    // State
    pdfUrl          string
    compileErrors   []string
    isLoading       bool
}
```

### Phase 3: Integration Features

#### 3.1 Real-time PDF Updates
```go
// Backend: Event-driven PDF compilation
func (mc *MusicComposer) handleDocumentChange(newLilyPond string) {
    // 1. Update document context
    // 2. Compile LilyPond to PDF
    // 3. Send SSE event: "pdf-ready"
    // 4. Frontend updates preview panel
}

// Frontend: EventSource listener
previewPanel.eventSource.AddEventListener("pdf-ready", func(event utils.EventSourceEvent) {
    previewPanel.updatePDF(event.Data.PDFUrl)
})
```

#### 3.2 Diff Visualization
```go
// Frontend: Render diff with color coding
func (ep *EditorPanel) renderDiffView() string {
    // Generate HTML with:
    // - Red background for deletions
    // - Green background for additions
    // - Yellow background for modifications
    // - Line numbers
    // - Context lines
}
```

#### 3.3 Context-Aware Prompts
```go
// Backend: Smart context management
func (pe *PromptEngine) buildContextualPrompt(userPrompt string) string {
    return fmt.Sprintf(`
You are a music composition assistant. You help users write musical notation in LilyPond.

Current document context:
- Key: %s
- Time signature: %s
- Current content: %s

Recent conversation:
%s

User request: %s

Please provide:
1. A clear explanation of your suggested changes
2. The complete updated LilyPond code
3. A diff showing exactly what changed

Format your response as JSON:
{
  "explanation": "...",
  "lilypond": "...",
  "diff": {
    "before": "...",
    "after": "...",
    "changes": [...]
  }
}
`, context.Key, context.TimeSignature, context.CurrentLilyPond, 
   formatConversationHistory(context.ConversationHistory), userPrompt)
}
```

### Phase 4: Advanced Features

#### 4.1 Document Versioning
```go
// Backend: Version control system
type VersionManager struct {
    versions map[string]*DocumentVersion
    current  string
}

func (vm *VersionManager) SaveVersion(description string) string {
    // Create new version with diff from current
    // Store in memory/database
    // Return version ID
}

func (vm *VersionManager) LoadVersion(id string) error {
    // Restore document to specific version
    // Update all panels
}
```

#### 4.2 Music-Specific Templates
```go
// Backend: Template system
type MusicTemplate struct {
    Name        string
    Description string
    LilyPond    string
    Metadata    MusicMetadata
}

var templates = map[string]MusicTemplate{
    "piano-sonata": {
        Name: "Piano Sonata",
        LilyPond: `\\version "2.24.0"
\\score {
    \\new PianoStaff <<
        \\new Staff { \\clef treble }
        \\new Staff { \\clef bass }
    >>
    \\layout {}
}`,
    },
    // More templates...
}
```

#### 4.3 Error Handling & Validation
```go
// Backend: LilyPond validation
type ValidationResult struct {
    IsValid    bool
    Errors     []LilyPondError
    Warnings   []LilyPondWarning
}

type LilyPondError struct {
    Line    int
    Column  int
    Message string
    Context string
}
```

## Implementation Timeline

### Week 1: Backend Foundation
- [ ] Extend anthropic-agent with context management
- [ ] Implement prompt engineering system
- [ ] Add diff generation capabilities
- [ ] Set up LilyPond processing pipeline

### Week 2: Frontend Foundation
- [ ] Create three-panel layout structure
- [ ] Implement chat panel with streaming
- [ ] Build basic LilyPond editor
- [ ] Add PDF preview panel

### Week 3: Integration & Polish
- [ ] Connect frontend to backend APIs
- [ ] Implement real-time PDF updates
- [ ] Add diff visualization
- [ ] Polish UI/UX

### Week 4: Advanced Features
- [ ] Document versioning system
- [ ] Music templates
- [ ] Error handling & validation
- [ ] Performance optimization

## Technical Considerations

### TinyGo Compatibility
- Use minimal external dependencies
- Avoid complex reflection
- Keep binary size under 2MB
- Use simple data structures

### Performance
- Stream responses for real-time feel
- Debounce LilyPond compilation
- Cache compiled PDFs
- Optimize diff generation

### Security
- Validate LilyPond input
- Sanitize user prompts
- Rate limit API calls
- Secure PDF serving

## File Structure

### Backend (./backends/music-composer)
```
backends/music-composer/
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── internal/
│   ├── agent/
│   │   └── music_agent.go
│   ├── context/
│   │   └── manager.go
│   ├── prompts/
│   │   └── engine.go
│   ├── diff/
│   │   └── generator.go
│   ├── lilypond/
│   │   └── processor.go
│   └── types/
│       └── types.go
└── README.md
```

### Frontend (./examples/musicComposer)
```
examples/musicComposer/
├── main.go
├── components/
│   ├── musicComposer/
│   │   ├── musicComposer.go
│   │   └── musicComposer.qtpl
│   ├── chatPanel/
│   │   ├── chatPanel.go
│   │   └── chatPanel.qtpl
│   ├── editorPanel/
│   │   ├── editorPanel.go
│   │   └── editorPanel.qtpl
│   └── previewPanel/
│       ├── previewPanel.go
│       └── previewPanel.qtpl
└── README.md
```

## Dependencies

### Backend Dependencies
- `github.com/anthropics/anthropic-sdk-go` - Anthropic API client
- `github.com/gorilla/mux` - HTTP router
- `github.com/google/uuid` - UUID generation
- `github.com/sergi/go-diff` - Diff generation

### Frontend Dependencies
- `github.com/cstevenson98/goFE/pkg/goFE` - Core framework
- `github.com/valyala/quicktemplate` - Template engine
- `github.com/google/uuid` - UUID generation

## API Specifications

### Chat Endpoints
```
POST /api/compose/chat
{
  "message": "string",
  "context": {
    "currentLilyPond": "string",
    "metadata": {...}
  }
}

Response:
{
  "success": true,
  "data": {
    "response": "string",
    "explanation": "string",
    "lilypond": "string",
    "diff": {...}
  }
}
```

### Streaming Endpoints
```
POST /api/compose/stream
{
  "message": "string",
  "context": {...}
}

Response:
{
  "success": true,
  "data": {
    "sessionId": "string"
  }
}

GET /api/compose/stream/{sessionId}
Server-Sent Events:
- event: message, data: "chunk"
- event: complete, data: "{}"
- event: error, data: "error message"
```

### Document Endpoints
```
POST /api/document/compile
{
  "lilypond": "string"
}

Response:
{
  "success": true,
  "data": {
    "pdfUrl": "string",
    "errors": ["string"]
  }
}

GET /api/document/pdf/{id}
Response: PDF file
```

## Testing Strategy

### Backend Testing
- Unit tests for each internal package
- Integration tests for API endpoints
- Mock Anthropic API responses
- LilyPond compilation testing

### Frontend Testing
- Component rendering tests
- Event handling tests
- Integration tests with backend
- UI/UX testing

### End-to-End Testing
- Complete workflow testing
- Performance testing
- Cross-browser compatibility
- Mobile responsiveness

## Deployment

### Development
- Local development with hot reload
- Docker Compose for full stack
- Environment variable configuration

### Production
- Docker containerization
- Reverse proxy (nginx)
- SSL/TLS termination
- Monitoring and logging

## Future Enhancements

### Phase 5: Advanced Music Features
- MIDI export/import
- Audio playback
- Music theory suggestions
- Collaborative editing

### Phase 6: AI Enhancements
- Music style analysis
- Harmonic progression suggestions
- Melody generation
- Arrangement optimization

### Phase 7: Platform Expansion
- Mobile app
- Desktop application
- Web browser extension
- API for third-party integrations 