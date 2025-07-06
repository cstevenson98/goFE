package agent

import (
	"context"
	"fmt"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

// AnthropicAgent defines the interface for interacting with Anthropic's Claude API
type AnthropicAgent interface {
	Instantiate() error
	SendMessage(message string) (string, error)
	SendMessageStream(message string) (io.Reader, error)
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

// streamingReader implements io.Reader for streaming responses
type streamingReader struct {
	stream    *ssestream.Stream[anthropic.MessageStreamEventUnion]
	message   anthropic.Message
	buffer    []byte
	bufferPos int
	done      bool
}

// Read implements io.Reader interface
func (sr *streamingReader) Read(p []byte) (n int, err error) {
	if sr.done {
		return 0, io.EOF
	}

	// If we have buffered data, return it
	if sr.bufferPos < len(sr.buffer) {
		n = copy(p, sr.buffer[sr.bufferPos:])
		sr.bufferPos += n
		return n, nil
	}

	// Get next event from stream
	if !sr.stream.Next() {
		if sr.stream.Err() != nil {
			return 0, fmt.Errorf("stream error: %w", sr.stream.Err())
		}
		sr.done = true
		return 0, io.EOF
	}

	event := sr.stream.Current()
	err = sr.message.Accumulate(event)
	if err != nil {
		return 0, fmt.Errorf("failed to accumulate message: %w", err)
	}

	// Extract text from content block delta events
	switch eventVariant := event.AsAny().(type) {
	case anthropic.ContentBlockDeltaEvent:
		switch deltaVariant := eventVariant.Delta.AsAny().(type) {
		case anthropic.TextDelta:
			sr.buffer = []byte(deltaVariant.Text)
			sr.bufferPos = 0
			n = copy(p, sr.buffer)
			sr.bufferPos = n
			return n, nil
		}
	}

	return 0, nil
}

// SendMessageStream sends a message to Claude and returns a streaming reader
func (a *anthropicAgent) SendMessageStream(message string) (io.Reader, error) {
	if a.client == nil {
		return nil, fmt.Errorf("agent not instantiated, call Instantiate() first")
	}

	ctx := context.Background()

	stream := a.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(message)),
		},
		Model: anthropic.ModelClaudeSonnet4_20250514,
	})

	return &streamingReader{
		stream:    stream,
		message:   anthropic.Message{},
		buffer:    make([]byte, 0),
		bufferPos: 0,
		done:      false,
	}, nil
}
