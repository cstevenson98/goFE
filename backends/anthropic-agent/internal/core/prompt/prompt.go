package prompt

import (
	"fmt"
	"strings"
)

// PromptEngine handles prompt engineering for the music composition assistant
type PromptEngine struct {
	systemPrompt    string
	lilypondVersion string
}

// NewPromptEngine creates a new prompt engine instance
func NewPromptEngine() *PromptEngine {
	pe := &PromptEngine{
		lilypondVersion: "2.22.1", // default version
	}
	pe.updateSystemPrompt()
	return pe
}

// NewPromptEngineWithVersion creates a new prompt engine instance with a specific LilyPond version
func NewPromptEngineWithVersion(version string) *PromptEngine {
	pe := &PromptEngine{
		lilypondVersion: version,
	}
	pe.updateSystemPrompt()
	return pe
}

// updateSystemPrompt updates the system prompt with the current LilyPond version
func (pe *PromptEngine) updateSystemPrompt() {
	pe.systemPrompt = fmt.Sprintf(`You are a music composition assistant that helps users create musical scores using LilyPond notation. You have expertise in:

1. Music theory and composition
2. LilyPond music notation syntax (version %s)
3. Score formatting and layout
4. Musical analysis and suggestions

Your role is to:
- Help users create musical scores in LilyPond format
- Provide guidance on music theory and composition
- Suggest improvements to musical pieces
- Explain musical concepts and notation
- Assist with score formatting and layout

When responding:
- Be clear and educational about musical concepts
- Provide LilyPond code examples when appropriate
- Always use version %s in your LilyPond code examples
- Explain your reasoning for musical suggestions
- Be encouraging and supportive of the user's musical journey

Always respond in a helpful, knowledgeable manner focused on music composition and LilyPond notation.`, pe.lilypondVersion, pe.lilypondVersion)
}

// GetSystemPrompt returns the system prompt
func (cp *PromptEngine) GetSystemPrompt() string {
	return cp.systemPrompt
}

// GetLilyPondVersion returns the current LilyPond version
func (cp *PromptEngine) GetLilyPondVersion() string {
	return cp.lilypondVersion
}

// SetLilyPondVersion updates the LilyPond version and regenerates the system prompt
func (cp *PromptEngine) SetLilyPondVersion(version string) {
	cp.lilypondVersion = version
	cp.updateSystemPrompt()
}

// BuildPrompt constructs a prompt from user input and context
func (cp *PromptEngine) BuildPrompt(userInput string, context map[string]interface{}) string {
	var prompt strings.Builder

	prompt.WriteString(cp.systemPrompt)
	prompt.WriteString("\n\n")

	// Add LilyPond content if provided
	if context != nil {
		if lilypondContent, exists := context["lilypond_content"]; exists && lilypondContent != "" {
			if content, ok := lilypondContent.(string); ok && content != "" {
				prompt.WriteString("Current LilyPond content to analyze or modify:\n")
				prompt.WriteString("```lilypond\n")
				prompt.WriteString(content)
				prompt.WriteString("\n```")
				prompt.WriteString("\n\n")
			}
		}
	}

	prompt.WriteString("User: ")
	prompt.WriteString(userInput)
	prompt.WriteString("\n\nAssistant: ")

	return prompt.String()
}

// EnhancePrompt enhances a prompt with additional context
func (cp *PromptEngine) EnhancePrompt(basePrompt string, enhancements []string) string {
	// Stub implementation - just concatenate for now
	if len(enhancements) == 0 {
		return basePrompt
	}

	var enhanced strings.Builder
	enhanced.WriteString(basePrompt)
	enhanced.WriteString("\n\nAdditional Context:\n")

	for _, enhancement := range enhancements {
		enhanced.WriteString("- ")
		enhanced.WriteString(enhancement)
		enhanced.WriteString("\n")
	}

	return enhanced.String()
}

// ValidatePrompt validates prompt content
func (cp *PromptEngine) ValidatePrompt(prompt string) error {
	// Stub implementation - basic validation
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if len(prompt) > 10000 {
		return fmt.Errorf("prompt too long (max 10000 characters)")
	}

	return nil
}

// SetSystemPrompt updates the system prompt
func (cp *PromptEngine) SetSystemPrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("system prompt cannot be empty")
	}

	cp.systemPrompt = prompt
	return nil
}
