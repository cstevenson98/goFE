package core

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/agent"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/core/lilypond"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/core/prompt"
)

// MusicAssistant integrates LilyPond processing with AI-powered composition assistance
type MusicAssistant struct {
	lilypondProcessor *lilypond.LilyPondProcessor
	promptEngine      *prompt.PromptEngine
	anthropicAgent    agent.AnthropicAgent
}

// MusicAssistantRequest represents a request to the music assistant
type MusicAssistantRequest struct {
	Message    string            `json:"message" validate:"required"`
	Content    string            `json:"content,omitempty"` // Current LilyPond content to analyze/modify
	Context    map[string]string `json:"context,omitempty"`
	DocumentID string            `json:"documentId,omitempty"`
}

// MusicAssistantResponse represents a response from the music assistant
type MusicAssistantResponse struct {
	Response      string                  `json:"response,omitempty"`
	NewContent    string                  `json:"newContent,omitempty"`
	CompileResult *lilypond.CompileResult `json:"compileResult,omitempty"`
	Analysis      string                  `json:"analysis,omitempty"`
	Suggestions   []string                `json:"suggestions,omitempty"`
	Message       string                  `json:"message"`
	Error         string                  `json:"error,omitempty"`
}

// NewMusicAssistant creates a new music assistant instance
func NewMusicAssistant(lp *lilypond.LilyPondProcessor, pe *prompt.PromptEngine, aa agent.AnthropicAgent) *MusicAssistant {
	return &MusicAssistant{
		lilypondProcessor: lp,
		promptEngine:      pe,
		anthropicAgent:    aa,
	}
}

// SendMessage sends a message to the music assistant and returns the response
func (ma *MusicAssistant) SendMessage(ctx context.Context, req *MusicAssistantRequest) (*MusicAssistantResponse, error) {
	// Build enhanced prompt with LilyPond context
	prompt := ma.buildEnhancedPrompt(req.Message, req.Content, req.Context)

	// Get AI response
	response, err := ma.anthropicAgent.SendMessage(prompt)
	if err != nil {
		return &MusicAssistantResponse{
			Error: fmt.Sprintf("Failed to get AI response: %v", err),
		}, nil
	}

	// Process the response to extract any LilyPond content
	result := ma.processResponse(response, req.DocumentID)

	return result, nil
}

// SendMessageStream sends a message to the music assistant and returns a streaming reader
func (ma *MusicAssistant) SendMessageStream(ctx context.Context, req *MusicAssistantRequest) (io.Reader, error) {
	// Build enhanced prompt with LilyPond context
	prompt := ma.buildEnhancedPrompt(req.Message, req.Content, req.Context)

	// Get streaming response
	reader, err := ma.anthropicAgent.SendMessageStream(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get streaming response: %w", err)
	}

	return reader, nil
}

// buildEnhancedPrompt builds a prompt with LilyPond processing capabilities
func (ma *MusicAssistant) buildEnhancedPrompt(message string, content string, context map[string]string) string {
	var prompt strings.Builder

	prompt.WriteString(ma.promptEngine.GetSystemPrompt())
	prompt.WriteString("\n\n")
	prompt.WriteString("You are a music composition assistant with LilyPond processing capabilities. ")
	prompt.WriteString("You can create, modify, analyze, and suggest improvements to musical pieces. ")
	prompt.WriteString("When you provide LilyPond code, it will be returned as new content for the user to review. ")
	prompt.WriteString("You can also analyze existing pieces and provide suggestions.\n\n")

	prompt.WriteString("User request: ")
	prompt.WriteString(message)
	prompt.WriteString("\n\n")

	if content != "" {
		prompt.WriteString("Current LilyPond content to analyze or modify:\n")
		prompt.WriteString("```lilypond\n")
		prompt.WriteString(content)
		prompt.WriteString("\n```")
		prompt.WriteString("\n\n")
	}

	if context != nil {
		prompt.WriteString("Additional context:\n")
		for key, value := range context {
			prompt.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Please respond appropriately. If you're creating or modifying LilyPond code, ")
	prompt.WriteString("provide the complete code that can be compiled. The user will review and compile it themselves. ")
	prompt.WriteString("If you're analyzing, provide detailed analysis. If you're suggesting improvements, ")
	prompt.WriteString("provide specific suggestions.\n\n")

	prompt.WriteString("Assistant: ")

	return prompt.String()
}

// processResponse processes the AI response to extract LilyPond content and other information
func (ma *MusicAssistant) processResponse(response string, documentID string) *MusicAssistantResponse {
	result := &MusicAssistantResponse{
		Response: response,
		Message:  "Response processed successfully",
	}

	// Extract LilyPond content if present
	lilypondContent := ma.extractLilyPondContent(response)
	if lilypondContent != "" {
		result.NewContent = lilypondContent
		result.Message = "LilyPond content generated successfully"
	}

	// Extract analysis if the response seems analytical
	if ma.isAnalysisResponse(response) {
		result.Analysis = response
	}

	// Extract suggestions if the response seems like suggestions
	if ma.isSuggestionResponse(response) {
		result.Suggestions = ma.parseSuggestions(response)
	}

	return result
}

// extractLilyPondContent extracts LilyPond code from AI response
func (ma *MusicAssistant) extractLilyPondContent(response string) string {
	// Look for code blocks
	if strings.Contains(response, "```lilypond") {
		start := strings.Index(response, "```lilypond")
		if start != -1 {
			start += len("```lilypond")
			end := strings.Index(response[start:], "```")
			if end != -1 {
				return strings.TrimSpace(response[start : start+end])
			}
		}
	}

	// Look for code blocks without language specification
	if strings.Contains(response, "```") {
		start := strings.Index(response, "```")
		if start != -1 {
			start += 3
			end := strings.Index(response[start:], "```")
			if end != -1 {
				content := strings.TrimSpace(response[start : start+end])
				// Check if it looks like LilyPond content
				if strings.Contains(content, "\\version") || strings.Contains(content, "\\score") {
					return content
				}
			}
		}
	}

	// If no code blocks found, check if the entire response looks like LilyPond
	if strings.Contains(response, "\\version") || strings.Contains(response, "\\score") {
		return strings.TrimSpace(response)
	}

	return ""
}

// isAnalysisResponse checks if the response seems like an analysis
func (ma *MusicAssistant) isAnalysisResponse(response string) bool {
	analysisKeywords := []string{
		"analysis", "analyze", "structure", "harmonic", "melodic",
		"rhythmic", "form", "key", "scale", "chord", "progression",
		"musical", "composition", "piece", "work",
	}

	lowerResponse := strings.ToLower(response)
	for _, keyword := range analysisKeywords {
		if strings.Contains(lowerResponse, keyword) {
			return true
		}
	}

	return false
}

// isSuggestionResponse checks if the response seems like suggestions
func (ma *MusicAssistant) isSuggestionResponse(response string) bool {
	suggestionKeywords := []string{
		"suggest", "suggestion", "improve", "enhance", "consider",
		"try", "could", "might", "recommend", "advise", "tip",
	}

	lowerResponse := strings.ToLower(response)
	for _, keyword := range suggestionKeywords {
		if strings.Contains(lowerResponse, keyword) {
			return true
		}
	}

	return false
}

// parseSuggestions parses suggestions from AI response
func (ma *MusicAssistant) parseSuggestions(response string) []string {
	var suggestions []string

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") || strings.HasPrefix(line, "*")) {
			// Remove bullet points and clean up
			line = strings.TrimPrefix(line, "-")
			line = strings.TrimPrefix(line, "•")
			line = strings.TrimPrefix(line, "*")
			line = strings.TrimSpace(line)
			if line != "" {
				suggestions = append(suggestions, line)
			}
		}
	}

	return suggestions
}

// ValidateAndCompile validates and compiles LilyPond content
func (ma *MusicAssistant) ValidateAndCompile(content string, documentID string) (*MusicAssistantResponse, error) {
	// Validate syntax first
	validation := ma.lilypondProcessor.ValidateSyntax(content)
	if !validation.IsValid {
		return &MusicAssistantResponse{
			Error: "LilyPond syntax validation failed",
		}, nil
	}

	// Compile the content
	compileResult, err := ma.lilypondProcessor.CompileToPDFWithID(content, documentID)
	if err != nil {
		return &MusicAssistantResponse{
			Error: fmt.Sprintf("Compilation failed: %v", err),
		}, nil
	}

	return &MusicAssistantResponse{
		NewContent:    content,
		CompileResult: compileResult,
		Message:       "Content validated and compiled successfully",
	}, nil
}
