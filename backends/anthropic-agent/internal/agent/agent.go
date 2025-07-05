package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// AnthropicAgent defines the interface for interacting with Anthropic's Claude API
type AnthropicAgent interface {
	Instantiate() error
	SendMessage(message string) (string, error)
}

// anthropicAgent implements the AnthropicAgent interface
type anthropicAgent struct {
	client *anthropic.Client
}

// NewAnthropicAgent creates a new instance of the Anthropic agent
func NewAnthropicAgent() AnthropicAgent {
	return &anthropicAgent{}
}

// Instantiate initializes the Anthropic client
func (a *anthropicAgent) Instantiate() error {
	client := anthropic.NewClient()

	a.client = &client
	return nil
}

// SendMessage sends a message to Claude and returns the response
func (a *anthropicAgent) SendMessage(message string) (string, error) {
	if a.client == nil {
		return "", fmt.Errorf("agent not instantiated, call Instantiate() first")
	}

	ctx := context.Background()

	// Create the message using the new API
	response, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(message)),
		},
		Model: anthropic.ModelClaudeSonnet4_20250514,
	})
	if err != nil {
		return "", fmt.Errorf("failed to send message to Claude: %w", err)
	}

	// Extract the response text
	if len(response.Content) > 0 && response.Content[0].Type == "text" {
		return response.Content[0].Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Claude")
}
