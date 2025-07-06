# LilyPond Processing System for Music Composer Assistant

## Overview
The LilyPond processing system handles compilation of LilyPond notation files directly to PDF, syntax validation, and error reporting. It provides a robust pipeline for converting musical notation written in LilyPond syntax into rendered PDF documents for preview.

## Architecture

### Core Components

#### 1. LilyPondProcessor Struct
```go
type LilyPondProcessor struct {
    tempDir         string
    lilypond        string
    maxCompileTime  time.Duration
    maxRetries      int
    outputManager   *OutputManager
}

type CompileOptions struct {
    OutputFormat string // "pdf", "png", "svg"
    DPI          int    // For image output
    PaperSize    string // "a4", "letter", etc.
    Margins      string // "1in", "2cm", etc.
    StaffSize    int    // Staff size in points
    IncludePaths []string // Additional include paths
}
```

#### 2. Compilation Result Structure
```go
type CompileResult struct {
    Success      bool              `json:"success"`
    OutputPath   string            `json:"outputPath"`
    OutputData   []byte            `json:"outputData,omitempty"`
    Errors       []CompileError    `json:"errors,omitempty"`
    Warnings     []CompileWarning  `json:"warnings,omitempty"`
    CompileTime  time.Duration     `json:"compileTime"`
    OutputSize   int64             `json:"outputSize"`
}

type CompileError struct {
    Line        int    `json:"line"`
    Column      int    `json:"column"`
    Message     string `json:"message"`
    Context     string `json:"context"`
    ErrorType   string `json:"errorType"` // "syntax", "missing_file", "lilypond", etc.
}

type CompileWarning struct {
    Line        int    `json:"line"`
    Column      int    `json:"column"`
    Message     string `json:"message"`
    Context     string `json:"context"`
    WarningType string `json:"warningType"`
}
```

## LilyPond Compilation Pipeline

### 1. Basic PDF Compilation
```go
func (lp *LilyPondProcessor) CompileToPDF(lilypond string) (*CompileResult, error) {
    startTime := time.Now()
    
    // Create temporary directory
    tempDir, err := lp.createTempDir()
    if err != nil {
        return nil, fmt.Errorf("failed to create temp directory: %w", err)
    }
    defer lp.cleanupTempDir(tempDir)
    
    // Write LilyPond to file
    lyFile := filepath.Join(tempDir, "score.ly")
    if err := lp.writeLilyPondFile(lyFile, lilypond); err != nil {
        return nil, fmt.Errorf("failed to write LilyPond file: %w", err)
    }
    
    // Run LilyPond compilation
    result, err := lp.runLilyPond(lyFile, tempDir)
    if err != nil {
        return nil, fmt.Errorf("compilation failed: %w", err)
    }
    
    result.CompileTime = time.Since(startTime)
    return result, nil
}

func (lp *LilyPondProcessor) runLilyPond(lyFile, tempDir string) (*CompileResult, error) {
    cmd := exec.Command(lp.lilypond, 
        "--pdf",
        "--output=" + tempDir,
        lyFile)
    
    cmd.Dir = tempDir
    cmd.Stdout = &bytes.Buffer{}
    cmd.Stderr = &bytes.Buffer{}
    
    // Set timeout
    ctx, cancel := context.WithTimeout(context.Background(), lp.maxCompileTime)
    defer cancel()
    cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
    
    err := cmd.Run()
    
    // Parse output for errors and warnings
    stdout := cmd.Stdout.(*bytes.Buffer).String()
    stderr := cmd.Stderr.(*bytes.Buffer).String()
    
    result := &CompileResult{}
    
    if err != nil {
        result.Success = false
        result.Errors = lp.parseCompileErrors(stderr)
        return result, err
    }
    
    // Check for PDF output
    pdfFile := filepath.Join(tempDir, "score.pdf")
    if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
        result.Success = false
        result.Errors = append(result.Errors, CompileError{
            Message:   "PDF file was not generated",
            ErrorType: "output_missing",
        })
        return result, fmt.Errorf("PDF file not generated")
    }
    
    // Read PDF data
    pdfData, err := os.ReadFile(pdfFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read PDF: %w", err)
    }
    
    result.Success = true
    result.OutputPath = pdfFile
    result.OutputData = pdfData
    result.OutputSize = int64(len(pdfData))
    result.Warnings = lp.parseCompileWarnings(stdout)
    
    return result, nil
}
```

### 2. Error and Warning Parsing
```go
func (lp *LilyPondProcessor) parseCompileErrors(stderr string) []CompileError {
    var errors []CompileError
    
    lines := strings.Split(stderr, "\n")
    for _, line := range lines {
        if strings.Contains(line, "error:") || strings.Contains(line, "fatal:") {
            error := lp.parseErrorLine(line)
            if error != nil {
                errors = append(errors, *error)
            }
        }
    }
    
    return errors
}

func (lp *LilyPondProcessor) parseErrorLine(line string) *CompileError {
    // Parse LilyPond error format: "error: ..."
    if strings.Contains(line, "error:") {
        parts := strings.SplitN(line, "error:", 2)
        if len(parts) >= 2 {
            return &CompileError{
                Message:   strings.TrimSpace(parts[1]),
                ErrorType: "lilypond",
                Context:   line,
            }
        }
    }
    
    // Parse line number format: "score.ly:123: ..."
    if strings.Contains(line, ".ly:") {
        parts := strings.SplitN(line, ":", 3)
        if len(parts) >= 3 {
            lineNumStr := parts[1]
            if lineNum, err := strconv.Atoi(lineNumStr); err == nil {
                return &CompileError{
                    Line:      lineNum,
                    Message:   strings.TrimSpace(parts[2]),
                    ErrorType: "syntax",
                    Context:   line,
                }
            }
        }
    }
    
    return nil
}

func (lp *LilyPondProcessor) parseCompileWarnings(stdout string) []CompileWarning {
    var warnings []CompileWarning
    
    lines := strings.Split(stdout, "\n")
    for _, line := range lines {
        if strings.Contains(line, "warning:") {
            warning := lp.parseWarningLine(line)
            if warning != nil {
                warnings = append(warnings, *warning)
            }
        }
    }
    
    return warnings
}

func (lp *LilyPondProcessor) parseWarningLine(line string) *CompileWarning {
    if strings.Contains(line, "warning:") {
        parts := strings.SplitN(line, "warning:", 2)
        if len(parts) >= 2 {
            return &CompileWarning{
                Message:     strings.TrimSpace(parts[1]),
                WarningType: "lilypond",
                Context:     line,
            }
        }
    }
    
    return nil
}
```

### 3. Basic LilyPond Document Wrapper
```go
func (lp *LilyPondProcessor) WrapLilyPondContent(content string) string {
    return fmt.Sprintf(`\\version "2.24.0"

\\paper {
    indent = 0\\mm
    line-width = 120\\mm
    oddHeaderMarkup = ##f
    evenHeaderMarkup = ##f
    oddFooterMarkup = ##f
    evenFooterMarkup = ##f
}

%s`, content)
}

func (lp *LilyPondProcessor) CreateBasicScore(notes string) string {
    return fmt.Sprintf(`\\version "2.24.0"

\\score {
    \\new Staff {
        %s
    }
    \\layout {}
}`, notes)
}
```

## Syntax Validation

### 1. Pre-compilation Validation
```go
func (lp *LilyPondProcessor) ValidateSyntax(lilypond string) *ValidationResult {
    result := &ValidationResult{
        IsValid: true,
        Errors:  []ValidationError{},
        Warnings: []ValidationWarning{},
    }
    
    // Check for basic LilyPond syntax
    lp.checkBasicSyntax(lilypond, result)
    
    // Check for score structure
    lp.checkScoreStructure(lilypond, result)
    
    // Check for common LilyPond errors
    lp.checkCommonErrors(lilypond, result)
    
    result.IsValid = len(result.Errors) == 0
    return result
}

type ValidationResult struct {
    IsValid  bool              `json:"isValid"`
    Errors   []ValidationError `json:"errors"`
    Warnings []ValidationWarning `json:"warnings"`
}

type ValidationError struct {
    Line    int    `json:"line"`
    Column  int    `json:"column"`
    Message string `json:"message"`
    Context string `json:"context"`
}

type ValidationWarning struct {
    Line    int    `json:"line"`
    Column  int    `json:"column"`
    Message string `json:"message"`
    Context string `json:"context"`
}

func (lp *LilyPondProcessor) checkBasicSyntax(lilypond string, result *ValidationResult) {
    lines := strings.Split(lilypond, "\n")
    
    for i, line := range lines {
        lineNum := i + 1
        
        // Check for unmatched braces
        if lp.hasUnmatchedBraces(line) {
            result.Errors = append(result.Errors, ValidationError{
                Line:    lineNum,
                Message: "Unmatched braces",
                Context: line,
            })
        }
        
        // Check for basic LilyPond commands
        if lp.hasInvalidCommand(line) {
            result.Errors = append(result.Errors, ValidationError{
                Line:    lineNum,
                Message: "Invalid LilyPond command",
                Context: line,
            })
        }
    }
}

func (lp *LilyPondProcessor) checkScoreStructure(lilypond string, result *ValidationResult) {
    // Check for score environment
    if strings.Contains(lilypond, "\\score") {
        if !strings.Contains(lilypond, "\\score{") {
            result.Errors = append(result.Errors, ValidationError{
                Message: "Invalid score syntax",
                Context: "Use \\score{...} for musical notation",
            })
        }
    } else {
        result.Warnings = append(result.Warnings, ValidationWarning{
            Message: "No score found",
            Context: "Add \\score{...} for musical notation",
        })
    }
    
    // Check for version declaration
    if !strings.Contains(lilypond, "\\version") {
        result.Warnings = append(result.Warnings, ValidationWarning{
            Message: "No version declaration",
            Context: "Consider adding \\version \"2.24.0\"",
        })
    }
}

func (lp *LilyPondProcessor) checkCommonErrors(lilypond string, result *ValidationResult) {
    // Check for common LilyPond syntax issues
    if strings.Contains(lilypond, "\\new Staff") && !strings.Contains(lilypond, "\\score") {
        result.Errors = append(result.Errors, ValidationError{
            Message: "Staff must be inside a score",
            Context: "Wrap \\new Staff in \\score{...}",
        })
    }
    
    // Check for missing layout
    if strings.Contains(lilypond, "\\score{") && !strings.Contains(lilypond, "\\layout") {
        result.Warnings = append(result.Warnings, ValidationWarning{
            Message: "No layout specified",
            Context: "Consider adding \\layout {} to score",
        })
    }
}
```

## File Management

### 1. Temporary Directory Management
```go
func (lp *LilyPondProcessor) createTempDir() (string, error) {
    tempDir, err := os.MkdirTemp(lp.tempDir, "lilypond-compile-")
    if err != nil {
        return "", err
    }
    
    return tempDir, nil
}

func (lp *LilyPondProcessor) cleanupTempDir(tempDir string) {
    if err := os.RemoveAll(tempDir); err != nil {
        log.Printf("Failed to cleanup temp directory %s: %v", tempDir, err)
    }
}

func (lp *LilyPondProcessor) writeLilyPondFile(filepath, content string) error {
    return os.WriteFile(filepath, []byte(content), 0644)
}
```

### 2. Output File Management
```go
type OutputManager struct {
    outputDir string
    maxFiles  int
    fileCache map[string]*CachedFile
    mutex     sync.RWMutex
}

type CachedFile struct {
    Path      string
    Data      []byte
    Created   time.Time
    AccessCount int
}

func (om *OutputManager) SaveOutput(id string, data []byte, format string) (string, error) {
    om.mutex.Lock()
    defer om.mutex.Unlock()
    
    filename := fmt.Sprintf("%s.%s", id, format)
    filepath := filepath.Join(om.outputDir, filename)
    
    if err := os.WriteFile(filepath, data, 0644); err != nil {
        return "", err
    }
    
    // Cache the file
    om.fileCache[id] = &CachedFile{
        Path:       filepath,
        Data:       data,
        Created:    time.Now(),
        AccessCount: 1,
    }
    
    // Cleanup old files if needed
    om.cleanupOldFiles()
    
    return filepath, nil
}

func (om *OutputManager) GetOutput(id string) (*CachedFile, bool) {
    om.mutex.RLock()
    defer om.mutex.RUnlock()
    
    cached, exists := om.fileCache[id]
    if exists {
        cached.AccessCount++
        return cached, true
    }
    
    return nil, false
}
```

## API Integration

### 1. REST API Endpoints
```go
// POST /api/lilypond/compile
type CompileRequest struct {
    LilyPond    string         `json:"lilypond"`
    Options     CompileOptions `json:"options,omitempty"`
}

type CompileResponse struct {
    Success bool          `json:"success"`
    Data    *CompileResult `json:"data,omitempty"`
    Error   string        `json:"error,omitempty"`
}

// POST /api/lilypond/validate
type ValidateRequest struct {
    LilyPond string `json:"lilypond"`
}

type ValidateResponse struct {
    Success bool              `json:"success"`
    Data    *ValidationResult `json:"data,omitempty"`
    Error   string            `json:"error,omitempty"`
}

// GET /api/lilypond/pdf/{id}
// Returns PDF file
```

### 2. Implementation
```go
func (lp *LilyPondProcessor) HandleCompileRequest(req *CompileRequest) (*CompileResponse, error) {
    // Wrap the lilypond content if needed
    lilypond := lp.WrapLilyPondContent(req.LilyPond)
    
    // Compile LilyPond
    result, err := lp.CompileToPDF(lilypond)
    if err != nil {
        return &CompileResponse{
            Success: false,
            Error:   fmt.Sprintf("Compilation error: %v", err),
        }, nil
    }
    
    return &CompileResponse{
        Success: result.Success,
        Data:    result,
    }, nil
}

func (lp *LilyPondProcessor) HandleValidateRequest(req *ValidateRequest) (*ValidateResponse, error) {
    result := lp.ValidateSyntax(req.LilyPond)
    
    return &ValidateResponse{
        Success: result.IsValid,
        Data:    result,
    }, nil
}
```

## Error Handling and Recovery

### 1. Compilation Retry Logic
```go
func (lp *LilyPondProcessor) CompileWithRetry(lilypond string, maxRetries int) (*CompileResult, error) {
    var lastError error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        result, err := lp.CompileToPDF(lilypond)
        if err == nil && result.Success {
            return result, nil
        }
        
        lastError = err
        
        // If it's a timeout or system error, retry
        if lp.isRetryableError(err) {
            log.Printf("Compilation attempt %d failed, retrying: %v", attempt+1, err)
            time.Sleep(time.Duration(attempt+1) * time.Second)
            continue
        }
        
        // If it's a LilyPond syntax error, don't retry
        break
    }
    
    return nil, fmt.Errorf("compilation failed after %d attempts: %w", maxRetries+1, lastError)
}

func (lp *LilyPondProcessor) isRetryableError(err error) bool {
    if err == nil {
        return false
    }
    
    errorStr := err.Error()
    
    // Retry on timeout, system errors, temporary failures
    retryablePatterns := []string{
        "timeout",
        "killed",
        "signal",
        "temporary",
        "resource",
    }
    
    for _, pattern := range retryablePatterns {
        if strings.Contains(strings.ToLower(errorStr), pattern) {
            return true
        }
    }
    
    return false
}
```

### 2. Error Recovery Strategies
```go
func (lp *LilyPondProcessor) SuggestFixes(errors []CompileError) []FixSuggestion {
    var suggestions []FixSuggestion
    
    for _, err := range errors {
        suggestion := lp.generateFixSuggestion(err)
        if suggestion != nil {
            suggestions = append(suggestions, *suggestion)
        }
    }
    
    return suggestions
}

type FixSuggestion struct {
    Error       CompileError `json:"error"`
    Suggestion  string       `json:"suggestion"`
    Code        string       `json:"code,omitempty"`
    Confidence  float64      `json:"confidence"`
}

func (lp *LilyPondProcessor) generateFixSuggestion(err CompileError) *FixSuggestion {
    switch err.ErrorType {
    case "syntax":
        return &FixSuggestion{
            Error:      err,
            Suggestion: "Check syntax on line " + strconv.Itoa(err.Line),
            Confidence: 0.7,
        }
    case "missing_score":
        return &FixSuggestion{
            Error:      err,
            Suggestion: "Wrap content in score block",
            Code:       "\\score { ... \\layout {} }",
            Confidence: 0.9,
        }
    default:
        return &FixSuggestion{
            Error:      err,
            Suggestion: "Review LilyPond syntax",
            Confidence: 0.5,
        }
    }
}
```

## Testing

### 1. Unit Tests
```go
func TestLilyPondProcessor_CompileToPDF(t *testing.T) {
    processor := NewLilyPondProcessor("/tmp", "lilypond", 30*time.Second, 3)
    
    lilypond := `\\version "2.24.0"
\\score {
    \\new Staff { c' d' e' f' }
    \\layout {}
}`
    
    result, err := processor.CompileToPDF(lilypond)
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.NotEmpty(t, result.OutputData)
    assert.Greater(t, result.OutputSize, int64(0))
}

func TestLilyPondProcessor_ValidateSyntax(t *testing.T) {
    processor := NewLilyPondProcessor("/tmp", "lilypond", 30*time.Second, 3)
    
    lilypond := `\\score {
    \\new Staff { c' d' e' f' }
    \\layout {}
}`
    
    result := processor.ValidateSyntax(lilypond)
    
    assert.True(t, result.IsValid)
    assert.Contains(t, result.Warnings[0].Message, "No version declaration")
}
```

### 2. Integration Tests
```go
func TestLilyPondProcessor_EndToEnd(t *testing.T) {
    processor := NewLilyPondProcessor("/tmp", "lilypond", 30*time.Second, 3)
    
    // Test complete workflow
    lilypondContent := `\\score { \\new Staff { c' d' e' f' } \\layout {} }`
    
    // Validate first
    validation := processor.ValidateSyntax(lilypondContent)
    assert.True(t, validation.IsValid)
    
    // Compile
    result, err := processor.CompileToPDF(processor.WrapLilyPondContent(lilypondContent))
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Performance Considerations

### 1. Caching
- Cache compiled PDFs by content hash
- Cache validation results
- Cache template renders

### 2. Parallel Processing
- Compile multiple documents concurrently
- Use worker pool for compilation tasks
- Background cleanup of temporary files

### 3. Resource Management
- Limit concurrent compilations
- Monitor disk space usage
- Cleanup old cached files

## Future Enhancements

### 1. Advanced Features
- Multiple output formats (PNG, SVG, MIDI)
- Custom paper sizes and margins
- Advanced LilyPond options (staff size, fonts, etc.)
- MIDI generation for playback

### 2. Performance Improvements
- Incremental compilation
- Pre-compiled templates
- Distributed compilation

### 3. Enhanced Error Handling
- Machine learning for error prediction
- Automatic fix suggestions
- Interactive error resolution

### 4. Music Theory Integration
- Automatic chord analysis
- Scale detection
- Harmonic analysis
- Voice leading suggestions 