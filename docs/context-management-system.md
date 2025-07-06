# Context Management System

## Overview
The Context Management System is responsible for maintaining the state of the music composition document, including LaTeX content, conversation history, document versions, and music metadata. It provides a centralized way to track changes and maintain context for AI interactions.

## Architecture

### Core Types

```go
// DocumentContext represents the complete state of a music composition document
type DocumentContext struct {
    ID                  string                    `json:"id"`
    CurrentLaTeX        string                    `json:"currentLaTeX"`
    DocumentHistory     []DocumentVersion         `json:"documentHistory"`
    MusicMetadata       MusicMetadata             `json:"musicMetadata"`
    ConversationHistory []ConversationTurn        `json:"conversationHistory"`
    CreatedAt           time.Time                 `json:"createdAt"`
    UpdatedAt           time.Time                 `json:"updatedAt"`
    Version             int                       `json:"version"`
}

// MusicMetadata contains music-specific information about the composition
type MusicMetadata struct {
    Title           string   `json:"title"`
    Composer        string   `json:"composer"`
    Key             string   `json:"key"`
    TimeSignature   string   `json:"timeSignature"`
    Tempo           string   `json:"tempo"`
    Instruments     []string `json:"instruments"`
    Style           string   `json:"style"`
    Difficulty      string   `json:"difficulty"`
    Duration        string   `json:"duration"`
    Tags            []string `json:"tags"`
}

// DocumentVersion represents a saved version of the document
type DocumentVersion struct {
    ID          string       `json:"id"`
    Timestamp   time.Time    `json:"timestamp"`
    LaTeX       string       `json:"latex"`
    Description string       `json:"description"`
    Diff        *DiffResult  `json:"diff,omitempty"`
    Author      string       `json:"author"`
    Tags        []string     `json:"tags"`
    IsAutoSave  bool         `json:"isAutoSave"`
}

// ConversationTurn represents a single exchange between user and assistant
type ConversationTurn struct {
    ID                string       `json:"id"`
    UserPrompt        string       `json:"userPrompt"`
    AssistantResponse string       `json:"assistantResponse"`
    AppliedChanges    []Change     `json:"appliedChanges"`
    Timestamp         time.Time    `json:"timestamp"`
    ContextSnapshot   string       `json:"contextSnapshot"`
    TokensUsed        int          `json:"tokensUsed"`
    ProcessingTime    time.Duration `json:"processingTime"`
}

// Change represents a single modification to the document
type Change struct {
    ID          string      `json:"id"`
    Type        ChangeType  `json:"type"`
    LineNum     int         `json:"lineNum"`
    ColumnStart int         `json:"columnStart"`
    ColumnEnd   int         `json:"columnEnd"`
    OldText     string      `json:"oldText"`
    NewText     string      `json:"newText"`
    Context     string      `json:"context"`
    Description string      `json:"description"`
    Timestamp   time.Time   `json:"timestamp"`
}

// ChangeType represents the type of change made
type ChangeType string

const (
    ChangeTypeInsert  ChangeType = "INSERT"
    ChangeTypeDelete  ChangeType = "DELETE"
    ChangeTypeModify  ChangeType = "MODIFY"
    ChangeTypeReplace ChangeType = "REPLACE"
)

// DiffResult represents the result of comparing two versions
type DiffResult struct {
    Before     string      `json:"before"`
    After      string      `json:"after"`
    Changes    []Change    `json:"changes"`
    HunkRanges []HunkRange `json:"hunkRanges"`
    Summary    DiffSummary `json:"summary"`
}

// HunkRange represents a range of lines in a diff
type HunkRange struct {
    StartLine int `json:"startLine"`
    EndLine   int `json:"endLine"`
    Context   int `json:"context"`
}

// DiffSummary provides statistics about the diff
type DiffSummary struct {
    LinesAdded    int `json:"linesAdded"`
    LinesDeleted  int `json:"linesDeleted"`
    LinesModified int `json:"linesModified"`
    TotalChanges  int `json:"totalChanges"`
}
```

## Context Manager Interface

```go
// ContextManager defines the interface for managing document context
type ContextManager interface {
    // Core operations
    CreateContext(metadata MusicMetadata) (*DocumentContext, error)
    LoadContext(id string) (*DocumentContext, error)
    SaveContext(context *DocumentContext) error
    DeleteContext(id string) error
    
    // Version management
    SaveVersion(contextID string, description string, isAutoSave bool) (*DocumentVersion, error)
    LoadVersion(contextID string, versionID string) (*DocumentContext, error)
    ListVersions(contextID string) ([]DocumentVersion, error)
    DeleteVersion(contextID string, versionID string) error
    
    // Conversation management
    AddConversationTurn(contextID string, turn ConversationTurn) error
    GetConversationHistory(contextID string, limit int) ([]ConversationTurn, error)
    ClearConversationHistory(contextID string) error
    
    // Content operations
    UpdateLaTeX(contextID string, newLaTeX string) error
    GetCurrentLaTeX(contextID string) (string, error)
    ValidateLaTeX(latex string) (*ValidationResult, error)
    
    // Metadata operations
    UpdateMetadata(contextID string, metadata MusicMetadata) error
    GetMetadata(contextID string) (*MusicMetadata, error)
    
    // Search and filtering
    SearchContexts(query string) ([]DocumentContext, error)
    FilterByMetadata(filters map[string]string) ([]DocumentContext, error)
    
    // Export and import
    ExportContext(contextID string) ([]byte, error)
    ImportContext(data []byte) (*DocumentContext, error)
}
```

## Implementation Algorithm

### 1. Context Creation Algorithm
```go
func (cm *contextManager) CreateContext(metadata MusicMetadata) (*DocumentContext, error) {
    // 1. Validate metadata
    if err := cm.validateMetadata(metadata); err != nil {
        return nil, fmt.Errorf("invalid metadata: %w", err)
    }
    
    // 2. Generate unique ID
    id := cm.generateContextID()
    
    // 3. Create initial LaTeX template
    initialLaTeX := cm.generateInitialLaTeX(metadata)
    
    // 4. Create context object
    context := &DocumentContext{
        ID:                  id,
        CurrentLaTeX:        initialLaTeX,
        DocumentHistory:     make([]DocumentVersion, 0),
        MusicMetadata:       metadata,
        ConversationHistory: make([]ConversationTurn, 0),
        CreatedAt:           time.Now(),
        UpdatedAt:           time.Now(),
        Version:             1,
    }
    
    // 5. Save initial version
    initialVersion := &DocumentVersion{
        ID:          cm.generateVersionID(),
        Timestamp:   time.Now(),
        LaTeX:       initialLaTeX,
        Description: "Initial version",
        Author:      "system",
        IsAutoSave:  false,
    }
    context.DocumentHistory = append(context.DocumentHistory, *initialVersion)
    
    // 6. Persist to storage
    if err := cm.storage.Save(context); err != nil {
        return nil, fmt.Errorf("failed to save context: %w", err)
    }
    
    return context, nil
}
```

### 2. Version Management Algorithm
```go
func (cm *contextManager) SaveVersion(contextID string, description string, isAutoSave bool) (*DocumentVersion, error) {
    // 1. Load current context
    context, err := cm.LoadContext(contextID)
    if err != nil {
        return nil, fmt.Errorf("failed to load context: %w", err)
    }
    
    // 2. Generate diff from previous version
    var diff *DiffResult
    if len(context.DocumentHistory) > 0 {
        previousVersion := context.DocumentHistory[len(context.DocumentHistory)-1]
        diff = cm.generateDiff(previousVersion.LaTeX, context.CurrentLaTeX)
    }
    
    // 3. Create new version
    version := &DocumentVersion{
        ID:          cm.generateVersionID(),
        Timestamp:   time.Now(),
        LaTeX:       context.CurrentLaTeX,
        Description: description,
        Diff:        diff,
        Author:      "user", // TODO: Get from session
        IsAutoSave:  isAutoSave,
    }
    
    // 4. Add to history
    context.DocumentHistory = append(context.DocumentHistory, *version)
    context.Version++
    context.UpdatedAt = time.Now()
    
    // 5. Persist changes
    if err := cm.storage.Save(context); err != nil {
        return nil, fmt.Errorf("failed to save version: %w", err)
    }
    
    return version, nil
}
```

### 3. Conversation Management Algorithm
```go
func (cm *contextManager) AddConversationTurn(contextID string, turn ConversationTurn) error {
    // 1. Load context
    context, err := cm.LoadContext(contextID)
    if err != nil {
        return fmt.Errorf("failed to load context: %w", err)
    }
    
    // 2. Generate context snapshot
    turn.ContextSnapshot = cm.generateContextSnapshot(context)
    
    // 3. Apply changes if any
    if len(turn.AppliedChanges) > 0 {
        if err := cm.applyChanges(context, turn.AppliedChanges); err != nil {
            return fmt.Errorf("failed to apply changes: %w", err)
        }
    }
    
    // 4. Add to conversation history
    context.ConversationHistory = append(context.ConversationHistory, turn)
    
    // 5. Trim history if too long
    if len(context.ConversationHistory) > cm.maxConversationHistory {
        context.ConversationHistory = context.ConversationHistory[len(context.ConversationHistory)-cm.maxConversationHistory:]
    }
    
    // 6. Update timestamps
    context.UpdatedAt = time.Now()
    
    // 7. Persist changes
    if err := cm.storage.Save(context); err != nil {
        return fmt.Errorf("failed to save conversation: %w", err)
    }
    
    return nil
}
```

### 4. Context Snapshot Generation Algorithm
```go
func (cm *contextManager) generateContextSnapshot(context *DocumentContext) string {
    // 1. Extract key information
    snapshot := map[string]interface{}{
        "metadata": context.MusicMetadata,
        "latexLength": len(context.CurrentLaTeX),
        "version": context.Version,
        "conversationTurns": len(context.ConversationHistory),
        "lastModified": context.UpdatedAt,
    }
    
    // 2. Add recent conversation summary
    if len(context.ConversationHistory) > 0 {
        recentTurns := context.ConversationHistory[max(0, len(context.ConversationHistory)-5):]
        snapshot["recentConversation"] = recentTurns
    }
    
    // 3. Add document structure analysis
    structure := cm.analyzeDocumentStructure(context.CurrentLaTeX)
    snapshot["structure"] = structure
    
    // 4. Serialize to JSON
    data, _ := json.Marshal(snapshot)
    return string(data)
}
```

### 5. Document Structure Analysis Algorithm
```go
func (cm *contextManager) analyzeDocumentStructure(latex string) map[string]interface{} {
    structure := map[string]interface{}{
        "sections": make([]string, 0),
        "instruments": make([]string, 0),
        "measures": 0,
        "complexity": "unknown",
    }
    
    lines := strings.Split(latex, "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        // Count measures
        if strings.Contains(line, "\\bar") {
            structure["measures"] = structure["measures"].(int) + 1
        }
        
        // Extract instruments
        if strings.Contains(line, "\\new Staff") {
            if instrument := cm.extractInstrument(line); instrument != "" {
                instruments := structure["instruments"].([]string)
                structure["instruments"] = append(instruments, instrument)
            }
        }
        
        // Extract sections
        if strings.Contains(line, "\\section") || strings.Contains(line, "\\movement") {
            if section := cm.extractSection(line); section != "" {
                sections := structure["sections"].([]string)
                structure["sections"] = append(sections, section)
            }
        }
    }
    
    // Determine complexity
    structure["complexity"] = cm.determineComplexity(latex)
    
    return structure
}
```

## Storage Interface

```go
// Storage defines the interface for persisting context data
type Storage interface {
    Save(context *DocumentContext) error
    Load(id string) (*DocumentContext, error)
    Delete(id string) error
    List() ([]string, error)
    Search(query string) ([]DocumentContext, error)
    Close() error
}

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
    contexts map[string]*DocumentContext
    mutex    sync.RWMutex
}

// FileStorage implements file-based storage
type FileStorage struct {
    basePath string
    mutex    sync.RWMutex
}

// DatabaseStorage implements database storage
type DatabaseStorage struct {
    db   *sql.DB
    mutex sync.RWMutex
}
```

## Configuration

```go
// Config holds configuration for the context manager
type Config struct {
    MaxConversationHistory int           `json:"maxConversationHistory"`
    AutoSaveInterval       time.Duration `json:"autoSaveInterval"`
    MaxContextSize         int64         `json:"maxContextSize"`
    StorageType            string        `json:"storageType"`
    StorageConfig          StorageConfig `json:"storageConfig"`
    BackupEnabled          bool          `json:"backupEnabled"`
    BackupInterval         time.Duration `json:"backupInterval"`
}

// StorageConfig holds storage-specific configuration
type StorageConfig struct {
    FilePath    string `json:"filePath,omitempty"`
    DatabaseURL string `json:"databaseUrl,omitempty"`
    TablePrefix string `json:"tablePrefix,omitempty"`
}
```

## Error Handling

```go
// ContextError represents errors specific to context management
type ContextError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

const (
    ErrContextNotFound     = "CONTEXT_NOT_FOUND"
    ErrInvalidMetadata     = "INVALID_METADATA"
    ErrStorageError        = "STORAGE_ERROR"
    ErrVersionNotFound     = "VERSION_NOT_FOUND"
    ErrInvalidLaTeX        = "INVALID_LATEX"
    ErrContextTooLarge     = "CONTEXT_TOO_LARGE"
    ErrConversationLimit   = "CONVERSATION_LIMIT"
)

func (ce *ContextError) Error() string {
    return fmt.Sprintf("[%s] %s", ce.Code, ce.Message)
}
```

## Usage Examples

### Creating a New Context
```go
manager := NewContextManager(config)

metadata := MusicMetadata{
    Title:         "Piano Sonata in C Major",
    Composer:      "John Doe",
    Key:           "C major",
    TimeSignature: "4/4",
    Tempo:         "120 BPM",
    Instruments:   []string{"piano"},
}

context, err := manager.CreateContext(metadata)
if err != nil {
    log.Fatal(err)
}
```

### Adding a Conversation Turn
```go
turn := ConversationTurn{
    ID:                uuid.New().String(),
    UserPrompt:        "Add a crescendo to the second measure",
    AssistantResponse: "I'll add a crescendo marking...",
    AppliedChanges:    []Change{...},
    Timestamp:         time.Now(),
    TokensUsed:        150,
    ProcessingTime:    2 * time.Second,
}

err := manager.AddConversationTurn(contextID, turn)
```

### Saving a Version
```go
version, err := manager.SaveVersion(contextID, "Added crescendo to second measure", false)
if err != nil {
    log.Fatal(err)
}
```

## Performance Considerations

1. **Memory Management**: Implement LRU cache for frequently accessed contexts
2. **Lazy Loading**: Load conversation history only when needed
3. **Compression**: Compress LaTeX content for storage
4. **Indexing**: Index contexts by metadata for fast search
5. **Background Processing**: Handle auto-save and cleanup in background goroutines

## Security Considerations

1. **Input Validation**: Validate all user inputs
2. **Access Control**: Implement user-based access control
3. **Data Sanitization**: Sanitize LaTeX content
4. **Audit Trail**: Log all context modifications
5. **Encryption**: Encrypt sensitive metadata 