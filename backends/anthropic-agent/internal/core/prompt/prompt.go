package prompt

import (
	"fmt"
	"strings"
)

// PromptEngine handles prompt engineering for the music composition assistant
type PromptEngine struct {
	systemPrompt string
}

// NewPromptEngine creates a new prompt engine instance
func NewPromptEngine() *PromptEngine {
	return &PromptEngine{
		systemPrompt: `You are a music composition assistant that helps users create musical scores using LilyPond notation. You have expertise in:

1. Music theory and composition
2. LilyPond music notation syntax
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
- Explain your reasoning for musical suggestions
- Be encouraging and supportive of the user's musical journey

Always respond in a helpful, knowledgeable manner focused on music composition and LilyPond notation.`,
	}
}

// GetSystemPrompt returns the system prompt
func (cp *PromptEngine) GetSystemPrompt() string {
	return cp.systemPrompt
}

// BuildPrompt constructs a prompt from user input and context
func (cp *PromptEngine) BuildPrompt(userInput string, context map[string]interface{}) string {
	// Stub implementation - just return system prompt + user input for now
	var prompt strings.Builder

	prompt.WriteString(cp.systemPrompt)
	prompt.WriteString("\n\nUser: ")
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

// GetPromptTemplate returns a prompt template by name
func (cp *PromptEngine) GetPromptTemplate(templateName string) (string, error) {
	// Stub implementation - return basic template
	templates := map[string]string{
		"composition": "Create a musical composition in LilyPond notation for: ",
		"analysis":    "Analyze this musical piece: ",
		"improvement": "Suggest improvements for this composition: ",
		"explanation": "Explain this musical concept: ",
	}

	if template, exists := templates[templateName]; exists {
		return template, nil
	}

	return "", fmt.Errorf("template not found: %s", templateName)
}

// SetSystemPrompt updates the system prompt
func (cp *PromptEngine) SetSystemPrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("system prompt cannot be empty")
	}

	cp.systemPrompt = prompt
	return nil
}
