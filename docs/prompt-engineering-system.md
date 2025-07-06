# Prompt Engineering System for Music Composer Assistant

## Overview
The prompt engineering system is a critical component that transforms user requests into structured, context-aware prompts for the Anthropic Claude API. It ensures consistent, high-quality responses for music composition tasks while maintaining conversation context and document state.

## Architecture

### Core Components

#### 1. PromptEngine Struct
```go
type PromptEngine struct {
    baseSystemPrompt string
    contextWindow    int
    maxTokens        int
    musicTemplates   map[string]string
    conversationBuffer *ConversationBuffer
}
```

#### 2. ConversationBuffer
```go
type ConversationBuffer struct {
    maxTurns     int
    turns        []ConversationTurn
    totalTokens  int
}

type ConversationTurn struct {
    Role        string    // "user" or "assistant"
    Content     string
    Timestamp   time.Time
    TokenCount  int
}
```

## System Prompts

### Base System Prompt
```go
const baseSystemPrompt = `You are an expert music composition assistant specializing in LilyPond notation. Your role is to help users create, modify, and refine musical scores.

CORE CAPABILITIES:
- Write and edit LilyPond code for musical notation
- Provide clear explanations of musical concepts
- Generate diffs showing exact changes made
- Maintain musical consistency and theory accuracy
- Suggest improvements and alternatives

RESPONSE FORMAT:
Always respond with valid JSON in this exact structure:
{
  "explanation": "Clear explanation of what you're doing and why",
  "lilypond": "Complete LilyPond code (not just the changes)",
  "diff": {
    "before": "Previous LilyPond content",
    "after": "New LilyPond content", 
    "changes": [
      {
        "type": "INSERT|DELETE|MODIFY",
        "line": 5,
        "oldText": "\\clef treble",
        "newText": "\\clef bass",
        "context": "Changed clef for bass line"
      }
    ]
  },
  "metadata": {
    "key": "C major",
    "timeSignature": "4/4",
    "tempo": "120",
    "instruments": ["piano"]
  }
}

MUSICAL GUIDELINES:
- Use proper LilyPond syntax
- Include appropriate clefs, time signatures, and key signatures
- Maintain consistent voice leading
- Follow standard musical notation conventions
- Provide context for complex musical decisions`
```

### Context-Aware Prompt Builder
```go
func (pe *PromptEngine) BuildMusicCompositionPrompt(
    userPrompt string, 
    context *DocumentContext,
    includeDiff bool,
) string {
    var prompt strings.Builder
    
    // 1. System instructions
    prompt.WriteString(pe.baseSystemPrompt)
    prompt.WriteString("\n\n")
    
    // 2. Current document context
    if context != nil {
        prompt.WriteString("CURRENT DOCUMENT CONTEXT:\n")
        prompt.WriteString(fmt.Sprintf("- Key: %s\n", context.MusicMetadata.Key))
        prompt.WriteString(fmt.Sprintf("- Time Signature: %s\n", context.MusicMetadata.TimeSignature))
        prompt.WriteString(fmt.Sprintf("- Tempo: %s\n", context.MusicMetadata.Tempo))
        prompt.WriteString(fmt.Sprintf("- Instruments: %s\n", strings.Join(context.MusicMetadata.Instruments, ", ")))
        prompt.WriteString("\n")
        
        // Current LilyPond content (truncated if too long)
        if len(context.CurrentLilyPond) > 1000 {
            prompt.WriteString("Current LilyPond (truncated):\n")
            prompt.WriteString(context.CurrentLilyPond[:1000])
            prompt.WriteString("\n... [truncated]\n\n")
        } else {
            prompt.WriteString("Current LilyPond:\n")
            prompt.WriteString(context.CurrentLilyPond)
            prompt.WriteString("\n\n")
        }
    }
    
    // 3. Recent conversation history
    if len(context.ConversationHistory) > 0 {
        prompt.WriteString("RECENT CONVERSATION:\n")
        recentTurns := pe.getRecentConversationTurns(context.ConversationHistory, 5)
        for _, turn := range recentTurns {
            prompt.WriteString(fmt.Sprintf("User: %s\n", turn.UserPrompt))
            prompt.WriteString(fmt.Sprintf("Assistant: %s\n", turn.AssistantResponse))
            prompt.WriteString("---\n")
        }
        prompt.WriteString("\n")
    }
    
    // 4. User's specific request
    prompt.WriteString("USER REQUEST:\n")
    prompt.WriteString(userPrompt)
    prompt.WriteString("\n\n")
    
    // 5. Special instructions based on request type
    prompt.WriteString(pe.getSpecialInstructions(userPrompt))
    
    return prompt.String()
}
```

## Prompt Templates

### Template System
```go
type PromptTemplate struct {
    Name        string
    Description string
    Template    string
    Variables   []string
}

var musicTemplates = map[string]PromptTemplate{
    "create-score": {
        Name: "Create New Score",
        Template: `Create a new musical score with the following specifications:
- Key: {{.Key}}
- Time Signature: {{.TimeSignature}}
- Tempo: {{.Tempo}}
- Instruments: {{.Instruments}}
- Style: {{.Style}}

Please provide a complete LilyPond document with proper notation.`,
        Variables: []string{"Key", "TimeSignature", "Tempo", "Instruments", "Style"},
    },
    "modify-measure": {
        Name: "Modify Specific Measure",
        Template: `Modify measure {{.MeasureNumber}} in the current score:
- Current content: {{.CurrentContent}}
- Requested change: {{.ChangeRequest}}

Provide the updated LilyPond with the specific measure modified.`,
        Variables: []string{"MeasureNumber", "CurrentContent", "ChangeRequest"},
    },
    "add-harmony": {
        Name: "Add Harmonic Progression",
        Template: `Add harmonic progression to the current melody:
- Current melody: {{.Melody}}
- Desired harmony: {{.HarmonyType}}
- Key: {{.Key}}

Create a complete score with melody and harmony.`,
        Variables: []string{"Melody", "HarmonyType", "Key"},
    },
}
```

### Template Rendering
```go
func (pe *PromptEngine) RenderTemplate(templateName string, data map[string]interface{}) (string, error) {
    template, exists := musicTemplates[templateName]
    if !exists {
        return "", fmt.Errorf("template %s not found", templateName)
    }
    
    tmpl, err := template.New(templateName).Parse(template.Template)
    if err != nil {
        return "", err
    }
    
    var result strings.Builder
    err = tmpl.Execute(&result, data)
    if err != nil {
        return "", err
    }
    
    return result.String(), nil
}
```

## Context Management

### Conversation History Management
```go
func (pe *PromptEngine) AddConversationTurn(role, content string) {
    turn := ConversationTurn{
        Role:       role,
        Content:    content,
        Timestamp:  time.Now(),
        TokenCount: pe.estimateTokenCount(content),
    }
    
    pe.conversationBuffer.turns = append(pe.conversationBuffer.turns, turn)
    pe.conversationBuffer.totalTokens += turn.TokenCount
    
    // Trim if buffer is too large
    pe.trimConversationBuffer()
}

func (pe *PromptEngine) trimConversationBuffer() {
    maxTokens := pe.contextWindow * 3 / 4 // Keep 75% of context window
    
    for pe.conversationBuffer.totalTokens > maxTokens && len(pe.conversationBuffer.turns) > 1 {
        removed := pe.conversationBuffer.turns[0]
        pe.conversationBuffer.turns = pe.conversationBuffer.turns[1:]
        pe.conversationBuffer.totalTokens -= removed.TokenCount
    }
}
```

### Document Context Integration
```go
func (pe *PromptEngine) BuildDocumentContextPrompt(context *DocumentContext) string {
    var prompt strings.Builder
    
    prompt.WriteString("DOCUMENT STATE:\n")
    
    // Metadata
    if context.MusicMetadata.Title != "" {
        prompt.WriteString(fmt.Sprintf("Title: %s\n", context.MusicMetadata.Title))
    }
    prompt.WriteString(fmt.Sprintf("Composer: %s\n", context.MusicMetadata.Composer))
    prompt.WriteString(fmt.Sprintf("Key: %s\n", context.MusicMetadata.Key))
    prompt.WriteString(fmt.Sprintf("Time Signature: %s\n", context.MusicMetadata.TimeSignature))
    prompt.WriteString(fmt.Sprintf("Tempo: %s\n", context.MusicMetadata.Tempo))
    prompt.WriteString(fmt.Sprintf("Instruments: %s\n", strings.Join(context.MusicMetadata.Instruments, ", ")))
    
    // Recent changes
    if len(context.DocumentHistory) > 0 {
        prompt.WriteString("\nRECENT CHANGES:\n")
        recentChanges := context.DocumentHistory[len(context.DocumentHistory)-3:]
        for _, change := range recentChanges {
            prompt.WriteString(fmt.Sprintf("- %s: %s\n", 
                change.Timestamp.Format("15:04:05"), 
                change.Description))
        }
    }
    
    return prompt.String()
}
```

## Specialized Prompt Types

### 1. Error Recovery Prompts
```go
func (pe *PromptEngine) BuildErrorRecoveryPrompt(errorMsg string, context *DocumentContext) string {
    return fmt.Sprintf(`The LilyPond compilation failed with the following error:
%s

Please analyze the error and provide a corrected version of the LilyPond code. 
Focus on fixing the specific syntax or structural issues mentioned in the error.

Current LilyPond:
%s

Provide the corrected LilyPond with an explanation of what was wrong and how you fixed it.`, 
        errorMsg, context.CurrentLilyPond)
}
```

### 2. Style Analysis Prompts
```go
func (pe *PromptEngine) BuildStyleAnalysisPrompt(context *DocumentContext) string {
    return fmt.Sprintf(`Analyze the musical style and characteristics of the current composition:

Current LilyPond:
%s

Please provide:
1. Musical style identification (Classical, Jazz, Pop, etc.)
2. Key harmonic features
3. Rhythmic patterns
4. Melodic characteristics
5. Suggestions for stylistic improvements

Format your response as JSON with these fields.`, context.CurrentLilyPond)
}
```

### 3. Theory Explanation Prompts
```go
func (pe *PromptEngine) BuildTheoryExplanationPrompt(concept string, context *DocumentContext) string {
    return fmt.Sprintf(`Explain the musical theory concept "%s" in the context of the current composition:

Current LilyPond:
%s

Provide:
1. Clear explanation of the concept
2. How it applies to this specific piece
3. Examples from the current notation
4. Suggestions for applying this concept

Make the explanation accessible to musicians of varying skill levels.`, concept, context.CurrentLilyPond)
}
```

## Token Management

### Token Counting
```go
func (pe *PromptEngine) estimateTokenCount(text string) int {
    // Rough estimation: 1 token ≈ 4 characters for English text
    return len(text) / 4
}

func (pe *PromptEngine) GetPromptTokenCount(prompt string) int {
    return pe.estimateTokenCount(prompt)
}

func (pe *PromptEngine) IsWithinTokenLimit(prompt string) bool {
    return pe.GetPromptTokenCount(prompt) <= pe.maxTokens
}
```

### Prompt Optimization
```go
func (pe *PromptEngine) OptimizePrompt(prompt string) string {
    if pe.IsWithinTokenLimit(prompt) {
        return prompt
    }
    
    // Truncate conversation history
    lines := strings.Split(prompt, "\n")
    var optimizedLines []string
    
    // Keep system prompt and current context
    systemEnd := 0
    for i, line := range lines {
        if strings.Contains(line, "USER REQUEST:") {
            systemEnd = i
            break
        }
        optimizedLines = append(optimizedLines, line)
    }
    
    // Add truncated user request
    optimizedLines = append(optimizedLines, "USER REQUEST:")
    optimizedLines = append(optimizedLines, pe.truncateToFit(lines[systemEnd+1:]))
    
    return strings.Join(optimizedLines, "\n")
}

func (pe *PromptEngine) truncateToFit(lines []string) []string {
    var result []string
    currentTokens := 0
    maxTokens := pe.maxTokens / 2 // Reserve half for response
    
    for _, line := range lines {
        lineTokens := pe.estimateTokenCount(line)
        if currentTokens + lineTokens > maxTokens {
            break
        }
        result = append(result, line)
        currentTokens += lineTokens
    }
    
    if len(result) < len(lines) {
        result = append(result, "... [truncated]")
    }
    
    return result
}
```

## Response Parsing

### JSON Response Parser
```go
type MusicCompositionResponse struct {
    Explanation string     `json:"explanation"`
    LilyPond    string     `json:"lilypond"`
    Diff        *DiffData  `json:"diff"`
    Metadata    *Metadata  `json:"metadata"`
}

type DiffData struct {
    Before  string    `json:"before"`
    After   string    `json:"after"`
    Changes []Change  `json:"changes"`
}

type Change struct {
    Type    string `json:"type"`
    Line    int    `json:"line"`
    OldText string `json:"oldText"`
    NewText string `json:"newText"`
    Context string `json:"context"`
}

type Metadata struct {
    Key           string   `json:"key"`
    TimeSignature string   `json:"timeSignature"`
    Tempo         string   `json:"tempo"`
    Instruments   []string `json:"instruments"`
}

func (pe *PromptEngine) ParseResponse(responseText string) (*MusicCompositionResponse, error) {
    // Clean up response text
    cleaned := pe.cleanResponseText(responseText)
    
    var response MusicCompositionResponse
    err := json.Unmarshal([]byte(cleaned), &response)
    if err != nil {
        return nil, fmt.Errorf("failed to parse JSON response: %w", err)
    }
    
    return &response, nil
}

func (pe *PromptEngine) cleanResponseText(text string) string {
    // Remove markdown code blocks
    text = strings.TrimPrefix(text, "```json")
    text = strings.TrimPrefix(text, "```")
    text = strings.TrimSuffix(text, "```")
    
    // Remove leading/trailing whitespace
    text = strings.TrimSpace(text)
    
    return text
}
```

## Usage Examples

### Basic Usage
```go
engine := NewPromptEngine(4000, 8000) // 4K context, 8K max tokens

// Build a simple composition prompt
prompt := engine.BuildMusicCompositionPrompt(
    "Create a simple C major scale in treble clef",
    nil, // No existing context
    false,
)

// Send to API and parse response
response, err := engine.ParseResponse(apiResponse)
if err != nil {
    log.Printf("Failed to parse response: %v", err)
    return
}

fmt.Printf("Explanation: %s\n", response.Explanation)
fmt.Printf("LilyPond: %s\n", response.LilyPond)
```

### Context-Aware Usage
```go
context := &DocumentContext{
    CurrentLilyPond: `\score { \new Staff { c' d' e' f' } \layout {} }`,
    MusicMetadata: MusicMetadata{
        Key: "C major",
        TimeSignature: "4/4",
        Tempo: "120",
        Instruments: []string{"piano"},
    },
    ConversationHistory: []ConversationTurn{
        {
            UserPrompt: "Create a simple melody",
            AssistantResponse: "I've created a C major scale...",
            Timestamp: time.Now().Add(-5 * time.Minute),
        },
    },
}

prompt := engine.BuildMusicCompositionPrompt(
    "Add harmony to this melody",
    context,
    true, // Include diff
)
```

### Template Usage
```go
data := map[string]interface{}{
    "Key": "C major",
    "TimeSignature": "4/4", 
    "Tempo": "120",
    "Instruments": []string{"piano"},
    "Style": "classical",
}

prompt, err := engine.RenderTemplate("create-score", data)
if err != nil {
    log.Printf("Failed to render template: %v", err)
    return
}
```

## Testing

### Unit Tests
```go
func TestPromptEngine_BuildMusicCompositionPrompt(t *testing.T) {
    engine := NewPromptEngine(4000, 8000)
    
    context := &DocumentContext{
        CurrentLilyPond: "test lilypond",
        MusicMetadata: MusicMetadata{Key: "C major"},
    }
    
    prompt := engine.BuildMusicCompositionPrompt("test request", context, false)
    
    // Assertions
    assert.Contains(t, prompt, "test request")
    assert.Contains(t, prompt, "C major")
    assert.Contains(t, prompt, "test lilypond")
}

func TestPromptEngine_ParseResponse(t *testing.T) {
    engine := NewPromptEngine(4000, 8000)
    
    responseText := `{
        "explanation": "test explanation",
        "lilypond": "test lilypond",
        "diff": {"before": "old", "after": "new", "changes": []},
        "metadata": {"key": "C major"}
    }`
    
    response, err := engine.ParseResponse(responseText)
    
    assert.NoError(t, err)
    assert.Equal(t, "test explanation", response.Explanation)
    assert.Equal(t, "test lilypond", response.LilyPond)
    assert.Equal(t, "C major", response.Metadata.Key)
}
```

## Performance Considerations

### Caching
- Cache frequently used templates
- Cache parsed responses
- Cache token count estimates

### Optimization
- Pre-compile templates
- Use string builders for concatenation
- Minimize memory allocations

### Monitoring
- Track prompt token usage
- Monitor response parsing success rates
- Log prompt generation times

## Future Enhancements

### 1. Dynamic Prompt Templates
- User-customizable templates
- Template versioning
- A/B testing for prompt effectiveness

### 2. Advanced Context Management
- Semantic context compression
- Intelligent context pruning
- Multi-document context

### 3. Response Validation
- LilyPond syntax validation
- Musical theory validation
- Response quality scoring

### 4. Prompt Optimization
- Machine learning for prompt effectiveness
- Automatic prompt tuning
- Performance-based prompt selection 