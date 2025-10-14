package test

import (
	"fmt"
	"testing"
	"time"
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
}

// CanvasError represents a canvas-related error
type CanvasError struct {
	Message string
}

func (e *CanvasError) Error() string {
	return e.Message
}

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

func TestMockCanvasManager_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		canvasID    string
		expectError bool
		expectInit  bool
	}{
		{
			name:        "WebGPU initialization",
			canvasID:    "test-webgpu",
			expectError: false,
			expectInit:  true,
		},
		{
			name:        "WebGL fallback",
			canvasID:    "test-webgl",
			expectError: false,
			expectInit:  true,
		},
		{
			name:        "Error scenario",
			canvasID:    "test-error",
			expectError: true,
			expectInit:  false,
		},
		{
			name:        "Default initialization",
			canvasID:    "default-canvas",
			expectError: false,
			expectInit:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockCanvasManager()

			err := mock.Initialize(tt.canvasID)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			initialized, _ := mock.GetStatus()
			if initialized != tt.expectInit {
				t.Errorf("Expected initialized=%v, got %v", tt.expectInit, initialized)
			}
		})
	}
}

func TestMockCanvasManager_Render(t *testing.T) {
	mock := NewMockCanvasManager()

	// Test render before initialization
	err := mock.Render()
	if err == nil {
		t.Errorf("Expected error when rendering before initialization")
	}

	// Initialize and test render
	mock.Initialize("test-canvas")

	err = mock.Render()
	if err != nil {
		t.Errorf("Expected no error when rendering after initialization, got: %v", err)
	}

	if mock.GetRenderCount() != 1 {
		t.Errorf("Expected render count 1, got %d", mock.GetRenderCount())
	}

	// Test multiple renders
	mock.Render()
	mock.Render()

	if mock.GetRenderCount() != 3 {
		t.Errorf("Expected render count 3, got %d", mock.GetRenderCount())
	}
}

func TestMockCanvasManager_Cleanup(t *testing.T) {
	mock := NewMockCanvasManager()

	// Initialize first
	mock.Initialize("test-canvas")

	// Test cleanup
	err := mock.Cleanup()
	if err != nil {
		t.Errorf("Expected no error during cleanup, got: %v", err)
	}

	if !mock.WasCleanupCalled() {
		t.Errorf("Expected cleanup to be called")
	}

	initialized, _ := mock.GetStatus()
	if initialized {
		t.Errorf("Expected canvas to be uninitialized after cleanup")
	}
}

func TestMockCanvasManager_Status(t *testing.T) {
	mock := NewMockCanvasManager()

	// Test initial status
	initialized, message := mock.GetStatus()
	if initialized {
		t.Errorf("Expected initial status to be uninitialized")
	}
	if message != "" {
		t.Errorf("Expected initial message to be empty, got: %s", message)
	}

	// Test status after initialization
	mock.Initialize("test-canvas")
	initialized, message = mock.GetStatus()
	if !initialized {
		t.Errorf("Expected status to be initialized after init")
	}
	if message == "" {
		t.Errorf("Expected non-empty message after initialization")
	}

	// Test status update
	mock.SetStatus(false, "Test error")
	initialized, message = mock.GetStatus()
	if initialized {
		t.Errorf("Expected status to be uninitialized after SetStatus")
	}
	if message != "Test error" {
		t.Errorf("Expected message 'Test error', got: %s", message)
	}
}

func TestCanvasError(t *testing.T) {
	err := &CanvasError{Message: "Test error"}
	if err.Error() != "Test error" {
		t.Errorf("Expected error message 'Test error', got: %s", err.Error())
	}
}
