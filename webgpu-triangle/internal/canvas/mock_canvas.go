package canvas

import (
	"fmt"
	"time"

	"github.com/conor/webgpu-triangle/internal/types"
)

// MockCanvasManager implements CanvasManager for testing
type MockCanvasManager struct {
	initialized   bool
	error         string
	renderCount   int
	cleanupCalled bool
}

// NewMockCanvasManager creates a new mock canvas manager
func NewMockCanvasManager() *MockCanvasManager {
	return &MockCanvasManager{
		initialized:   false,
		error:         "",
		renderCount:   0,
		cleanupCalled: false,
	}
}

// Initialize simulates canvas initialization
func (m *MockCanvasManager) Initialize(canvasID string) error {
	fmt.Printf("Mock: Initializing canvas with ID: %s\n", canvasID)

	// Simulate initialization delay
	time.Sleep(100 * time.Millisecond)

	// Simulate different scenarios based on canvas ID
	switch canvasID {
	case "test-webgpu":
		m.initialized = true
		m.error = "Mock WebGPU triangle rendered successfully!"
		fmt.Println("Mock: WebGPU initialization successful")
	case "test-webgl":
		m.initialized = true
		m.error = "Mock WebGL triangle rendered successfully!"
		fmt.Println("Mock: WebGL fallback successful")
	case "test-error":
		m.initialized = false
		m.error = "Mock initialization failed"
		return &CanvasError{Message: "Mock initialization failed"}
	default:
		m.initialized = true
		m.error = "Mock canvas initialized successfully!"
		fmt.Println("Mock: Default initialization successful")
	}

	return nil
}

// Render simulates rendering a frame
func (m *MockCanvasManager) Render() error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	m.renderCount++
	fmt.Printf("Mock: Rendering frame #%d\n", m.renderCount)
	return nil
}

// Cleanup simulates resource cleanup
func (m *MockCanvasManager) Cleanup() error {
	m.cleanupCalled = true
	m.initialized = false
	m.error = "Mock cleanup completed"
	fmt.Println("Mock: Cleanup called")
	return nil
}

// GetStatus returns the current status
func (m *MockCanvasManager) GetStatus() (bool, string) {
	return m.initialized, m.error
}

// SetStatus updates the status
func (m *MockCanvasManager) SetStatus(initialized bool, message string) {
	m.initialized = initialized
	m.error = message
	fmt.Printf("Mock: Status updated - initialized: %v, message: %s\n", initialized, message)
}

// GetRenderCount returns the number of times Render was called
func (m *MockCanvasManager) GetRenderCount() int {
	return m.renderCount
}

// WasCleanupCalled returns whether Cleanup was called
func (m *MockCanvasManager) WasCleanupCalled() bool {
	return m.cleanupCalled
}

// New rendering methods for mock

// DrawTexture draws a texture at the specified position and size
func (m *MockCanvasManager) DrawTexture(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Printf("Mock: DrawTexture - Position: (%.2f, %.2f), Size: (%.2f, %.2f), UV: (%.2f, %.2f, %.2f, %.2f)\n",
		position.X, position.Y, size.X, size.Y, uv.U, uv.V, uv.W, uv.H)
	return nil
}

// DrawTextureRotated draws a rotated texture
func (m *MockCanvasManager) DrawTextureRotated(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, rotation float64) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Printf("Mock: DrawTextureRotated - Position: (%.2f, %.2f), Rotation: %.2f degrees\n",
		position.X, position.Y, rotation*180.0/3.14159)
	return nil
}

// DrawTextureScaled draws a scaled texture
func (m *MockCanvasManager) DrawTextureScaled(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, scale types.Vector2) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Printf("Mock: DrawTextureScaled - Position: (%.2f, %.2f), Scale: (%.2f, %.2f)\n",
		position.X, position.Y, scale.X, scale.Y)
	return nil
}

// BeginBatch starts batch rendering mode
func (m *MockCanvasManager) BeginBatch() error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Println("Mock: BeginBatch called")
	return nil
}

// EndBatch ends batch rendering mode
func (m *MockCanvasManager) EndBatch() error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Println("Mock: EndBatch called")
	return nil
}

// FlushBatch forces rendering of batched vertices
func (m *MockCanvasManager) FlushBatch() error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Println("Mock: FlushBatch called")
	return nil
}

// GetSpritePipeline returns the sprite rendering pipeline
func (m *MockCanvasManager) GetSpritePipeline() types.Pipeline {
	return &types.WebGPUPipeline{Valid: true}
}

// GetBackgroundPipeline returns the background rendering pipeline
func (m *MockCanvasManager) GetBackgroundPipeline() types.Pipeline {
	return &types.WebGPUPipeline{Valid: true}
}

// ClearCanvas clears the canvas
func (m *MockCanvasManager) ClearCanvas() error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Println("Mock: ClearCanvas called")
	return nil
}

// SetPipelines sets the active rendering pipelines
func (m *MockCanvasManager) SetPipelines(pipelines []types.PipelineType) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Printf("Mock: SetPipelines called with %d pipelines\n", len(pipelines))
	for i, p := range pipelines {
		fmt.Printf("  Pipeline %d: %s\n", i, p.String())
	}
	return nil
}

// DrawColoredRect draws a colored rectangle
func (m *MockCanvasManager) DrawColoredRect(position types.Vector2, size types.Vector2, color [4]float32) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Printf("Mock: DrawColoredRect - Position: (%.2f, %.2f), Size: (%.2f, %.2f), Color: (%.2f, %.2f, %.2f, %.2f)\n",
		position.X, position.Y, size.X, size.Y, color[0], color[1], color[2], color[3])
	return nil
}

// LoadTexture loads a texture from a path
func (m *MockCanvasManager) LoadTexture(path string) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Printf("Mock: LoadTexture - Path: %s\n", path)
	return nil
}

// DrawTexturedRect draws a textured rectangle
func (m *MockCanvasManager) DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	fmt.Printf("Mock: DrawTexturedRect - Texture: %s, Position: (%.2f, %.2f), Size: (%.2f, %.2f), UV: (%.2f, %.2f, %.2f, %.2f)\n",
		texturePath, position.X, position.Y, size.X, size.Y, uv.U, uv.V, uv.W, uv.H)
	return nil
}
