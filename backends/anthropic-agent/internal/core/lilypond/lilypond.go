package lilypond

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LilyPondProcessor handles compilation of LilyPond notation files
type LilyPondProcessor struct {
	outputDir      string
	lilypond       string
	maxCompileTime time.Duration
	maxRetries     int
	outputManager  *OutputManager
	version        string
}

// CompileOptions represents compilation options
type CompileOptions struct {
	OutputFormat string   // "pdf", "png", "svg"
	DPI          int      // For image output
	PaperSize    string   // "a4", "letter", etc.
	Margins      string   // "1in", "2cm", etc.
	StaffSize    int      // Staff size in points
	IncludePaths []string // Additional include paths
}

// CompileResult represents the result of a compilation
type CompileResult struct {
	Success     bool             `json:"success"`
	OutputPath  string           `json:"outputPath"`
	OutputData  []byte           `json:"outputData,omitempty"`
	Errors      []CompileError   `json:"errors,omitempty"`
	Warnings    []CompileWarning `json:"warnings,omitempty"`
	CompileTime time.Duration    `json:"compileTime"`
	OutputSize  int64            `json:"outputSize"`
	Stdout      string           `json:"stdout,omitempty"`
	Stderr      string           `json:"stderr,omitempty"`
}

// CompileError represents a compilation error
type CompileError struct {
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Message   string `json:"message"`
	Context   string `json:"context"`
	ErrorType string `json:"errorType"` // "syntax", "missing_file", "lilypond", etc.
}

// CompileWarning represents a compilation warning
type CompileWarning struct {
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Message     string `json:"message"`
	Context     string `json:"context"`
	WarningType string `json:"warningType"`
}

// ValidationResult represents syntax validation result
type ValidationResult struct {
	IsValid  bool                `json:"isValid"`
	Errors   []ValidationError   `json:"errors"`
	Warnings []ValidationWarning `json:"warnings"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Context string `json:"context"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Context string `json:"context"`
}

// OutputManager manages output files
type OutputManager struct {
	outputDir string
	maxFiles  int
	fileCache map[string]*CachedFile
	mutex     sync.RWMutex
}

// CachedFile represents a cached file
type CachedFile struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"createdAt"`
	AccessedAt time.Time `json:"accessedAt"`
}

// NewLilyPondProcessor creates a new LilyPond processor instance
func NewLilyPondProcessor() *LilyPondProcessor {
	// Determine output directory from DOCUMENT_DIR env var or use current working directory
	outputDir := "."
	if docDir := os.Getenv("DOCUMENT_DIR"); docDir != "" {
		outputDir = docDir
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Warning: Failed to create output directory %s: %v", outputDir, err)
		outputDir = "."
	}

	processor := &LilyPondProcessor{
		outputDir:      outputDir,
		lilypond:       "lilypond",
		maxCompileTime: 30 * time.Second,
		maxRetries:     3,
		outputManager: &OutputManager{
			outputDir: outputDir,
			maxFiles:  100,
			fileCache: make(map[string]*CachedFile),
		},
	}

	// Get LilyPond version
	processor.version = processor.getLilyPondVersion()

	return processor
}

// CompileToPDF compiles LilyPond content to PDF
func (lp *LilyPondProcessor) CompileToPDF(lilypond string) (*CompileResult, error) {
	return lp.CompileToPDFWithID(lilypond, "")
}

// CompileToPDFWithID compiles LilyPond content to PDF with a specific document ID
func (lp *LilyPondProcessor) CompileToPDFWithID(lilypond string, documentID string) (*CompileResult, error) {
	startTime := time.Now()

	// Clean and validate the LilyPond content
	cleanedContent := lp.cleanLilyPondContent(lilypond)

	// Log the cleaned content for debugging
	log.Printf("Compiling LilyPond content for document %s:\n%s", documentID, cleanedContent)

	// Write LilyPond to file in the output directory
	lyFile := filepath.Join(lp.outputDir, documentID+".ly")
	if err := lp.writeLilyPondFile(lyFile, cleanedContent); err != nil {
		result := &CompileResult{
			Success: false,
			Errors: []CompileError{
				{
					Message:   fmt.Sprintf("Failed to write LilyPond file: %v", err),
					ErrorType: "file_write_error",
					Context:   lyFile,
				},
			},
		}
		return result, fmt.Errorf("failed to write LilyPond file: %w", err)
	}

	// Verify the file was created
	if _, err := os.Stat(lyFile); os.IsNotExist(err) {
		result := &CompileResult{
			Success: false,
			Errors: []CompileError{
				{
					Message:   fmt.Sprintf("LilyPond file was not created at %s", lyFile),
					ErrorType: "file_missing",
					Context:   lyFile,
				},
			},
		}
		return result, fmt.Errorf("LilyPond file was not created at %s", lyFile)
	}
	log.Printf("LilyPond file created successfully at: %s", lyFile)

	// Run LilyPond compilation in the output directory
	result, err := lp.runLilyPond(lyFile, lp.outputDir)
	if err != nil {
		// The result should already contain stdout/stderr from runLilyPond
		// Just add the compile time and return the result with the error
		if result != nil {
			result.CompileTime = time.Since(startTime)
		}
		return result, fmt.Errorf("compilation failed: %w", err)
	}

	result.CompileTime = time.Since(startTime)
	return result, nil
}

// runLilyPond executes the LilyPond compiler
func (lp *LilyPondProcessor) runLilyPond(lyFile, outputDir string) (*CompileResult, error) {
	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), lp.maxCompileTime)
	defer cancel()

	// Get just the filename from the full path
	lyFileName := filepath.Base(lyFile)
	baseName := strings.TrimSuffix(lyFileName, ".ly")

	// Specify the exact output filename
	outputFile := filepath.Join(outputDir, baseName)

	cmd := exec.CommandContext(ctx, lp.lilypond,
		"--pdf",
		"--output="+outputFile,
		lyFileName)

	cmd.Dir = outputDir
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	// Log the command being executed for debugging
	log.Printf("Executing LilyPond command: %s --pdf --output=%s %s (in dir: %s)",
		lp.lilypond, outputFile, lyFileName, outputDir)

	err := cmd.Run()

	// Parse output for errors and warnings
	stdout := cmd.Stdout.(*bytes.Buffer).String()
	stderr := cmd.Stderr.(*bytes.Buffer).String()

	result := &CompileResult{
		Stdout: stdout,
		Stderr: stderr,
	}

	// Always log the full stderr output for debugging
	if stderr != "" {
		log.Printf("LilyPond stderr output:\n%s", stderr)
	}

	if err != nil {
		result.Success = false
		result.Errors = lp.parseCompileErrors(stderr)

		// If no errors were parsed but we have stderr, create a generic error
		if len(result.Errors) == 0 && stderr != "" {
			result.Errors = append(result.Errors, CompileError{
				Message:   fmt.Sprintf("LilyPond compilation failed: %v", err),
				ErrorType: "compilation_failed",
				Context:   stderr,
			})
		}

		// Return a more detailed error message
		errorMsg := fmt.Sprintf("LilyPond compilation failed: %v", err)
		if stderr != "" {
			errorMsg += fmt.Sprintf("\nStderr: %s", stderr)
		}
		return result, fmt.Errorf(errorMsg)
	}

	// Check for PDF output - use the same base name as the input file
	pdfFile := filepath.Join(outputDir, baseName+".pdf")
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		result.Success = false
		result.Errors = append(result.Errors, CompileError{
			Message:   "PDF file was not generated",
			ErrorType: "output_missing",
			Context:   fmt.Sprintf("stdout: %s\nstderr: %s", stdout, stderr),
		})
		return result, fmt.Errorf("PDF file not generated. stdout: %s, stderr: %s", stdout, stderr)
	}

	// Read PDF data
	pdfData, err := os.ReadFile(pdfFile)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, CompileError{
			Message:   fmt.Sprintf("Failed to read PDF: %v", err),
			ErrorType: "pdf_read_error",
			Context:   pdfFile,
		})
		return result, fmt.Errorf("failed to read PDF: %w", err)
	}

	result.Success = true
	result.OutputPath = pdfFile
	result.OutputData = pdfData
	result.OutputSize = int64(len(pdfData))
	result.Warnings = lp.parseCompileWarnings(stdout)

	return result, nil
}

// parseCompileErrors parses compilation errors from stderr
func (lp *LilyPondProcessor) parseCompileErrors(stderr string) []CompileError {
	var errors []CompileError

	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for various error patterns
		if strings.Contains(line, "error:") ||
			strings.Contains(line, "fatal:") ||
			strings.Contains(line, "syntax error") ||
			strings.Contains(line, "not a note name") ||
			strings.Contains(line, "unexpected") {
			error := lp.parseErrorLine(line)
			if error != nil {
				errors = append(errors, *error)
			}
		}
	}

	return errors
}

// parseErrorLine parses a single error line
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

	// Parse fatal error format: "fatal: ..."
	if strings.Contains(line, "fatal:") {
		parts := strings.SplitN(line, "fatal:", 2)
		if len(parts) >= 2 {
			return &CompileError{
				Message:   strings.TrimSpace(parts[1]),
				ErrorType: "fatal",
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

	// Parse general syntax errors: "syntax error, unexpected STRING, expecting '.' or '='"
	if strings.Contains(line, "syntax error") {
		return &CompileError{
			Message:   strings.TrimSpace(line),
			ErrorType: "syntax",
			Context:   line,
		}
	}

	// Parse "not a note name" errors
	if strings.Contains(line, "not a note name") {
		return &CompileError{
			Message:   strings.TrimSpace(line),
			ErrorType: "syntax",
			Context:   line,
		}
	}

	// Parse "unexpected" errors
	if strings.Contains(line, "unexpected") {
		return &CompileError{
			Message:   strings.TrimSpace(line),
			ErrorType: "syntax",
			Context:   line,
		}
	}

	// If no specific pattern matches, return a generic error
	return &CompileError{
		Message:   strings.TrimSpace(line),
		ErrorType: "unknown",
		Context:   line,
	}
}

// parseCompileWarnings parses compilation warnings from stdout
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

// parseWarningLine parses a single warning line
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

// ValidateSyntax validates LilyPond syntax without compilation
func (lp *LilyPondProcessor) ValidateSyntax(lilypond string) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
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

// checkBasicSyntax checks basic LilyPond syntax
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

// checkScoreStructure checks score structure
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
			Context: fmt.Sprintf("Consider adding \\version \"%s\"", lp.version),
		})
	}
}

// checkCommonErrors checks for common LilyPond errors
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

// hasUnmatchedBraces checks for unmatched braces
func (lp *LilyPondProcessor) hasUnmatchedBraces(line string) bool {
	openBraces := strings.Count(line, "{")
	closeBraces := strings.Count(line, "}")
	return openBraces != closeBraces
}

// hasInvalidCommand checks for invalid commands
func (lp *LilyPondProcessor) hasInvalidCommand(line string) bool {
	// Basic check for invalid commands - this is a simplified implementation
	invalidCommands := []string{"\\invalid", "\\broken"}
	for _, cmd := range invalidCommands {
		if strings.Contains(line, cmd) {
			return true
		}
	}
	return false
}

// writeLilyPondFile writes LilyPond content to a file
func (lp *LilyPondProcessor) writeLilyPondFile(filepath, content string) error {
	// Fix double-escaped backslashes
	fixedContent := lp.fixEscapedBackslashes(content)

	// Log the content being written for debugging
	log.Printf("Writing LilyPond content to %s:\n%s", filepath, fixedContent)

	return os.WriteFile(filepath, []byte(fixedContent), 0644)
}

// fixEscapedBackslashes fixes double-escaped backslashes in LilyPond content
func (lp *LilyPondProcessor) fixEscapedBackslashes(content string) string {
	// Replace double backslashes with single backslashes
	// This handles cases where the content has been escaped for JSON
	fixed := strings.ReplaceAll(content, "\\\\", "\\")
	return fixed
}

// cleanLilyPondContent cleans and validates LilyPond content
func (lp *LilyPondProcessor) cleanLilyPondContent(content string) string {
	// Fix escaped backslashes
	cleaned := lp.fixEscapedBackslashes(content)

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Ensure the content ends with a newline
	if !strings.HasSuffix(cleaned, "\n") {
		cleaned += "\n"
	}

	return cleaned
}

// SaveLilyPondSource saves LilyPond content to a file in the output directory
func (lp *LilyPondProcessor) SaveLilyPondSource(documentID, content string) error {
	if documentID == "" {
		return fmt.Errorf("document ID is required")
	}

	lyFile := filepath.Join(lp.outputDir, documentID+".ly")
	return lp.writeLilyPondFile(lyFile, content)
}

// WrapLilyPondContent wraps LilyPond content with basic structure
func (lp *LilyPondProcessor) WrapLilyPondContent(content string) string {
	return fmt.Sprintf(`\version "%s"

\paper {
    indent = 0\mm
    line-width = 120\mm
    oddHeaderMarkup = ##f
    evenHeaderMarkup = ##f
    oddFooterMarkup = ##f
    evenFooterMarkup = ##f
}

%s`, lp.version, content)
}

// CreateBasicScore creates a basic score structure
func (lp *LilyPondProcessor) CreateBasicScore(notes string) string {
	return fmt.Sprintf(`\version "%s"

\score {
    \new Staff {
        %s
    }
    \layout {}
}`, lp.version, notes)
}

// GetOutputDir returns the output directory path
func (lp *LilyPondProcessor) GetOutputDir() string {
	return lp.outputDir
}

// GetVersion returns the detected LilyPond version
func (lp *LilyPondProcessor) GetVersion() string {
	return lp.version
}

// getLilyPondVersion gets the LilyPond version by running lilypond --version
func (lp *LilyPondProcessor) getLilyPondVersion() string {
	cmd := exec.Command(lp.lilypond, "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to get LilyPond version: %v", err)
		return "2.22.1" // fallback version
	}

	// Parse the version from output like "GNU LilyPond 2.22.1"
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "GNU LilyPond") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "LilyPond" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}

	log.Printf("Warning: Could not parse LilyPond version from output: %s", outputStr)
	return "2.22.1" // fallback version
}
