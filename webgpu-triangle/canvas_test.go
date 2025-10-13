package main

import (
	"testing"
)

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
