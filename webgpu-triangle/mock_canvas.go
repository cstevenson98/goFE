package main

import (
	"fmt"
	"time"
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
