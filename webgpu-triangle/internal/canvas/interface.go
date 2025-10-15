package canvas

import (
	"github.com/conor/webgpu-triangle/internal/types"
)

// CanvasManager defines the interface for managing canvas operations
type CanvasManager interface {
	// Initialize sets up the canvas and returns success status
	Initialize(canvasID string) error

	// Render draws the current frame
	Render() error

	// Cleanup releases resources
	Cleanup() error

	// GetStatus returns the current status
	GetStatus() (bool, string)

	// SetStatus updates the status
	SetStatus(initialized bool, message string)

	// Sprite rendering methods (stubs for future implementation)
	DrawTexture(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect) error
	DrawTextureRotated(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, rotation float64) error
	DrawTextureScaled(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, scale types.Vector2) error

	// Batch rendering (stubs for future implementation)
	BeginBatch() error
	EndBatch() error
	FlushBatch() error

	// Pipeline management (stubs for future implementation)
	GetSpritePipeline() types.Pipeline
	GetBackgroundPipeline() types.Pipeline

	// Canvas management
	ClearCanvas() error

	// Helper methods for testing/debugging
	DrawColoredRect(position types.Vector2, size types.Vector2, color [4]float32) error

	// Texture loading
	LoadTexture(path string) error
	DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error
}

// CanvasError represents a canvas-related error
type CanvasError struct {
	Message string
}

func (e *CanvasError) Error() string {
	return e.Message
}
