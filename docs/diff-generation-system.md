# Diff Generation System for Music Composer Assistant

## Overview
The diff generation system provides basic difference detection and visualization for text-based code changes. It enables users to see exactly what modifications were made to their documents, with simple highlighting and structured change tracking.

## Architecture

### Core Components

#### 1. DiffGenerator Struct
```go
type DiffGenerator struct {
    contextLines int
    maxDiffSize  int
    tempDir      string
}
```

#### 2. DiffResult Structure
```go
type DiffResult struct {
    Before     string      `json:"before"`
    After      string      `json:"after"`
    Changes    []Change    `json:"changes"`
    HunkRanges []HunkRange `json:"hunkRanges"`
    Metadata   DiffMetadata `json:"metadata"`
}

type Change struct {
    Type      ChangeType `json:"type"`
    LineNum   int        `json:"lineNum"`
    OldText   string     `json:"oldText"`
    NewText   string     `json:"newText"`
    StartPos  int        `json:"startPos"`
    EndPos    int        `json:"endPos"`
}

type ChangeType string

const (
    ChangeInsert ChangeType = "INSERT"
    ChangeDelete ChangeType = "DELETE"
    ChangeModify ChangeType = "MODIFY"
    ChangeEqual  ChangeType = "EQUAL"
)

type HunkRange struct {
    StartLine int `json:"startLine"`
    EndLine   int `json:"endLine"`
    Changes   []Change `json:"changes"`
}

type DiffMetadata struct {
    TotalChanges    int    `json:"totalChanges"`
    Insertions      int    `json:"insertions"`
    Deletions       int    `json:"deletions"`
    Modifications   int    `json:"modifications"`
    GenerationTime  string `json:"generationTime"`
    Algorithm       string `json:"algorithm"`
}
```

## Diff Algorithm

### Myers Diff Algorithm
```go
type DiffGenerator struct {
    contextLines int
    maxDiffSize  int
    tempDir      string
}

func (dg *DiffGenerator) GenerateDiff(before, after string) (*DiffResult, error) {
    beforeLines := strings.Split(before, "\n")
    afterLines := strings.Split(after, "\n")
    
    // Use blackboxed diff library (e.g., github.com/sergi/go-diff)
    diff := diffmatchpatch.New()
    diffs := diff.DiffMain(before, after, true)
    
    return dg.processDiffs(diffs, beforeLines, afterLines)
}

func (dg *DiffGenerator) processDiffs(diffs []diffmatchpatch.Diff, beforeLines, afterLines []string) (*DiffResult, error) {
    var changes []Change
    var hunkRanges []HunkRange
    currentHunk := &HunkRange{}
    
    lineNum := 1
    for _, d := range diffs {
        switch d.Type {
        case diffmatchpatch.DiffEqual:
            if len(currentHunk.Changes) > 0 {
                currentHunk.EndLine = lineNum - 1
                hunkRanges = append(hunkRanges, *currentHunk)
                currentHunk = &HunkRange{StartLine: lineNum}
            }
            lineNum += strings.Count(d.Text, "\n")
            
        case diffmatchpatch.DiffDelete:
            change := Change{
                Type:     ChangeDelete,
                LineNum:  lineNum,
                OldText:  d.Text,
                NewText:  "",
                StartPos: lineNum,
                EndPos:   lineNum + strings.Count(d.Text, "\n"),
            }
            changes = append(changes, change)
            currentHunk.Changes = append(currentHunk.Changes, change)
            lineNum += strings.Count(d.Text, "\n")
            
        case diffmatchpatch.DiffInsert:
            change := Change{
                Type:     ChangeInsert,
                LineNum:  lineNum,
                OldText:  "",
                NewText:  d.Text,
                StartPos: lineNum,
                EndPos:   lineNum + strings.Count(d.Text, "\n"),
            }
            changes = append(changes, change)
            currentHunk.Changes = append(currentHunk.Changes, change)
            lineNum += strings.Count(d.Text, "\n")
        }
    }
    
    // Add final hunk if it has changes
    if len(currentHunk.Changes) > 0 {
        currentHunk.EndLine = lineNum - 1
        hunkRanges = append(hunkRanges, *currentHunk)
    }
    
    return &DiffResult{
        Before:     strings.Join(beforeLines, "\n"),
        After:      strings.Join(afterLines, "\n"),
        Changes:    changes,
        HunkRanges: hunkRanges,
        Metadata:   dg.generateMetadata(changes),
    }, nil
}

func (dg *DiffGenerator) generateMetadata(changes []Change) DiffMetadata {
    metadata := DiffMetadata{}
    
    for _, change := range changes {
        metadata.TotalChanges++
        switch change.Type {
        case ChangeInsert:
            metadata.Insertions++
        case ChangeDelete:
            metadata.Deletions++
        case ChangeModify:
            metadata.Modifications++
        }
    }
    
    return metadata
}
```

## Diff API

### 1. REST API Endpoints
```go
// POST /api/diff/generate
type GenerateDiffRequest struct {
    Before     string `json:"before"`
    After      string `json:"after"`
    ContextLines int  `json:"contextLines,omitempty"`
}

type GenerateDiffResponse struct {
    Success bool       `json:"success"`
    Data    *DiffResult `json:"data,omitempty"`
    Error   string     `json:"error,omitempty"`
}

// POST /api/diff/render
type RenderDiffRequest struct {
    Diff       *DiffResult `json:"diff"`
    Format     string      `json:"format"` // "html", "json", "unified"
    Theme      string      `json:"theme,omitempty"` // "light", "dark"
}

type RenderDiffResponse struct {
    Success bool   `json:"success"`
    Data    string `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

### 2. Implementation
```go
func (dg *DiffGenerator) GenerateDiff(before, after string) (*DiffResult, error) {
    startTime := time.Now()
    
    result, err := dg.GenerateDiff(before, after)
    if err != nil {
        return nil, err
    }
    
    // Add metadata
    result.Metadata.GenerationTime = time.Since(startTime).String()
    result.Metadata.Algorithm = "Myers"
    
    return result, nil
}

func (dg *DiffGenerator) RenderDiff(diff *DiffResult, format, theme string) (string, error) {
    switch format {
    case "html":
        renderer := &HTMLDiffRenderer{theme: dg.getTheme(theme)}
        return renderer.RenderHTML(diff), nil
    case "json":
        data, err := json.MarshalIndent(diff, "", "  ")
        return string(data), err
    case "unified":
        return dg.renderUnified(diff), nil
    default:
        return "", fmt.Errorf("unsupported format: %s", format)
    }
}
```

## Performance Optimization

### 1. Diff Caching
```go
type DiffCache struct {
    cache map[string]*CachedDiff
    mutex sync.RWMutex
    maxSize int
}

type CachedDiff struct {
    Result     *DiffResult
    Created    time.Time
    AccessCount int
}

func (dc *DiffCache) Get(before, after string) (*DiffResult, bool) {
    key := dc.generateKey(before, after)
    
    dc.mutex.RLock()
    cached, exists := dc.cache[key]
    dc.mutex.RUnlock()
    
    if exists {
        cached.AccessCount++
        return cached.Result, true
    }
    
    return nil, false
}

func (dc *DiffCache) Set(before, after string, result *DiffResult) {
    key := dc.generateKey(before, after)
    
    dc.mutex.Lock()
    defer dc.mutex.Unlock()
    
    // Evict if cache is full
    if len(dc.cache) >= dc.maxSize {
        dc.evictOldest()
    }
    
    dc.cache[key] = &CachedDiff{
        Result:     result,
        Created:    time.Now(),
        AccessCount: 1,
    }
}

func (dc *DiffCache) generateKey(before, after string) string {
    hash := sha256.New()
    hash.Write([]byte(before))
    hash.Write([]byte(after))
    return hex.EncodeToString(hash.Sum(nil))
}
```

### 2. Incremental Diffing
```go
type IncrementalDiffGenerator struct {
    baseGenerator *DiffGenerator
    cache         *DiffCache
}

func (idg *IncrementalDiffGenerator) GenerateIncrementalDiff(
    baseVersion string,
    currentVersion string,
    previousDiff *DiffResult,
) (*DiffResult, error) {
    // Check if we can use previous diff as base
    if previousDiff != nil && previousDiff.After == baseVersion {
        // Generate diff from previous result to current
        return idg.baseGenerator.GenerateDiff(previousDiff.After, currentVersion)
    }
    
    // Fallback to full diff
    return idg.baseGenerator.GenerateDiff(baseVersion, currentVersion)
}
```

## Testing

### 1. Unit Tests
```go
func TestDiffGenerator_GenerateDiff(t *testing.T) {
    generator := NewDiffGenerator(3, 1000, "/tmp")
    
    before := `function example() {
    console.log("hello");
    return true;
}`
    
    after := `function example() {
    console.log("hello world");
    return true;
}`
    
    result, err := generator.GenerateDiff(before, after)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, 1, result.Metadata.Modifications)
    assert.Equal(t, 0, result.Metadata.Deletions)
    assert.Equal(t, 0, result.Metadata.Insertions)
}

func TestDiffGenerator_BasicText(t *testing.T) {
    generator := NewDiffGenerator(3, 1000, "/tmp")
    
    before := `line 1
line 2
line 3`
    
    after := `line 1
line 2 modified
line 3`
    
    result, err := generator.GenerateDiff(before, after)
    
    assert.NoError(t, err)
    assert.Equal(t, 1, result.Metadata.Modifications)
}
```

### 2. Performance Tests
```go
func BenchmarkDiffGenerator_LargeFile(b *testing.B) {
    generator := NewDiffGenerator(3, 10000, "/tmp")
    
    // Generate large text files
    before := generateLargeTextFile(1000)
    after := generateLargeTextFile(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := generator.GenerateDiff(before, after)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func generateLargeTextFile(lines int) string {
    var content strings.Builder
    
    for i := 0; i < lines; i++ {
        content.WriteString(fmt.Sprintf("line %d: some content here\n", i+1))
    }
    
    return content.String()
}
```

## Integration with Frontend

### 1. Real-time Diff Updates
```go
// Frontend component integration
type DiffViewer struct {
    id           uuid.UUID
    diffContent  string
    isUpdating   bool
    eventSource  *utils.EventSource
}

func (dv *DiffViewer) UpdateDiff(newDiff *DiffResult) {
    dv.isUpdating = true
    
    // Render diff to HTML
    renderer := &HTMLDiffRenderer{theme: DefaultTheme}
    html := renderer.RenderHTML(newDiff)
    
    // Update component
    dv.diffContent = html
    dv.isUpdating = false
    
    // Trigger re-render
    dv.triggerUpdate()
}

func (dv *DiffViewer) triggerUpdate() {
    // Send update event to frontend
    event := utils.Event{
        Type: "diff-updated",
        Data: map[string]interface{}{
            "id":   dv.id.String(),
            "html": dv.diffContent,
        },
    }
    
    dv.eventSource.SendEvent(event)
}
```

### 2. Diff Navigation
```go
type DiffNavigator struct {
    currentHunk int
    totalHunks  int
    hunks       []HunkRange
}

func (dn *DiffNavigator) NextHunk() *HunkRange {
    if dn.currentHunk < len(dn.hunks)-1 {
        dn.currentHunk++
        return &dn.hunks[dn.currentHunk]
    }
    return nil
}

func (dn *DiffNavigator) PreviousHunk() *HunkRange {
    if dn.currentHunk > 0 {
        dn.currentHunk--
        return &dn.hunks[dn.currentHunk]
    }
    return nil
}

func (dn *DiffNavigator) GoToHunk(index int) *HunkRange {
    if index >= 0 && index < len(dn.hunks) {
        dn.currentHunk = index
        return &dn.hunks[dn.currentHunk]
    }
    return nil
}
```

## Future Enhancements

### 1. Enhanced Diff Features
- Side-by-side diff view
- Syntax highlighting in diffs
- Code folding in diff view

### 2. Interactive Diff Features
- Inline diff editing
- Diff conflict resolution
- Diff annotation and comments

### 3. Performance Improvements
- Parallel diff processing
- Streaming diff generation
- Diff compression and optimization

### 4. Enhanced Visualization
- Side-by-side diff view
- Syntax highlighting in diffs
- Code folding in diff view 