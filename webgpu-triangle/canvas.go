package main

import (
	"syscall/js"
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

// WebGPUCanvasManager implements CanvasManager using WebGPU
type WebGPUCanvasManager struct {
	canvas      js.Value
	device      js.Value
	context     js.Value
	pipeline    js.Value
	renderFunc  js.Func
	initialized bool
	error       string
}

// NewWebGPUCanvasManager creates a new WebGPU canvas manager
func NewWebGPUCanvasManager() *WebGPUCanvasManager {
	return &WebGPUCanvasManager{
		initialized: false,
		error:       "",
	}
}

// Initialize sets up the WebGPU canvas
func (w *WebGPUCanvasManager) Initialize(canvasID string) error {
	println("DEBUG: WebGPUCanvasManager.Initialize called for canvas:", canvasID)

	// Get the canvas element
	canvas := js.Global().Get("document").Call("getElementById", canvasID)
	if canvas.IsUndefined() || canvas.IsNull() {
		err := "Canvas element not found"
		w.SetStatus(false, err)
		return &CanvasError{Message: err}
	}

	w.canvas = canvas
	println("DEBUG: Canvas element found")

	// Ensure canvas has proper size
	canvas.Set("width", 800)
	canvas.Set("height", 600)
	println("DEBUG: Canvas size set to 800x600")

	// Check if WebGPU is supported
	gpu := js.Global().Get("navigator").Get("gpu")
	if gpu.IsUndefined() {
		println("DEBUG: WebGPU not available, falling back to WebGL")
		w.SetStatus(false, "WebGPU not available, using WebGL fallback")
		w.initializeWebGL()
		return nil
	}

	println("DEBUG: WebGPU available, initializing...")
	w.SetStatus(false, "Initializing WebGPU...")

	// Request adapter
	println("DEBUG: Requesting WebGPU adapter")
	adapterPromise := gpu.Call("requestAdapter")

	// Add timeout fallback to WebGL (5 seconds)
	timeoutId := js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("DEBUG: WebGPU timeout after 5 seconds, falling back to WebGL")
		w.SetStatus(false, "WebGPU initialization timed out, using WebGL fallback")
		w.initializeWebGL()
		return nil
	}), 5000) // 5 second timeout

	adapterPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Clear the timeout since WebGPU is working
		js.Global().Call("clearTimeout", timeoutId)

		println("DEBUG: WebGPU adapter promise resolved")
		if len(args) > 0 && !args[0].IsNull() {
			adapter := args[0]
			println("DEBUG: WebGPU adapter obtained, requesting device")

			// Request device
			devicePromise := adapter.Call("requestDevice")
			devicePromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				println("DEBUG: WebGPU device promise resolved")
				if len(args) > 0 {
					device := args[0]
					println("DEBUG: WebGPU device obtained, setting up triangle")
					w.setupWebGPUTriangle(device)
				} else {
					println("DEBUG: WebGPU device is null")
					w.SetStatus(false, "WebGPU device is null, using WebGL fallback")
					w.initializeWebGL()
				}
				return nil
			}))

			// Handle device promise rejection
			devicePromise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				println("DEBUG: WebGPU device promise rejected:", args[0].Get("message").String())
				w.SetStatus(false, "WebGPU device failed: "+args[0].Get("message").String()+", using WebGL fallback")
				w.initializeWebGL()
				return nil
			}))
		} else {
			println("DEBUG: WebGPU adapter is null")
			w.SetStatus(false, "WebGPU adapter is null, using WebGL fallback")
			w.initializeWebGL()
		}
		return nil
	}))

	// Handle adapter rejection
	adapterPromise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Clear the timeout since we're handling the error
		js.Global().Call("clearTimeout", timeoutId)

		println("DEBUG: WebGPU adapter promise rejected:", args[0].Get("message").String())
		w.SetStatus(false, "WebGPU adapter failed: "+args[0].Get("message").String()+", using WebGL fallback")
		w.initializeWebGL()
		return nil
	}))

	return nil
}

// Render draws the current frame
func (w *WebGPUCanvasManager) Render() error {
	if w.renderFunc.IsUndefined() {
		return &CanvasError{Message: "Render function not initialized"}
	}

	w.renderFunc.Invoke()
	return nil
}

// Cleanup releases resources
func (w *WebGPUCanvasManager) Cleanup() error {
	if !w.renderFunc.IsUndefined() {
		w.renderFunc.Release()
	}
	w.SetStatus(false, "Cleaned up")
	return nil
}

// GetStatus returns the current status
func (w *WebGPUCanvasManager) GetStatus() (bool, string) {
	return w.initialized, w.error
}

// SetStatus updates the status
func (w *WebGPUCanvasManager) SetStatus(initialized bool, message string) {
	w.initialized = initialized
	w.error = message
	// Update status in UI
	statusElement := js.Global().Get("document").Call("getElementById", "status-text")
	if !statusElement.IsUndefined() && !statusElement.IsNull() {
		statusElement.Set("textContent", message)
	}

	// Update status indicator
	indicator := js.Global().Get("document").Call("getElementById", "status-indicator")
	if !indicator.IsUndefined() && !indicator.IsNull() {
		indicator.Set("className", "status-indicator status-"+getStatusType(initialized, message))
	}

	println("STATUS:", message)
}

// setupWebGPUTriangle configures WebGPU rendering
func (w *WebGPUCanvasManager) setupWebGPUTriangle(device js.Value) {
	println("DEBUG: Setting up WebGPU triangle")

	// Get canvas context
	context := w.canvas.Call("getContext", "webgpu")
	if context.IsUndefined() {
		println("DEBUG: Failed to get WebGPU context")
		w.SetStatus(false, "Failed to get WebGPU context, using WebGL fallback")
		w.initializeWebGL()
		return
	}

	w.context = context
	w.device = device

	// Configure canvas
	canvasFormat := js.Global().Get("navigator").Get("gpu").Get("getPreferredCanvasFormat").Call("call", js.Global().Get("navigator").Get("gpu"))
	context.Call("configure", js.ValueOf(map[string]interface{}{
		"device":    device,
		"format":    canvasFormat,
		"alphaMode": "premultiplied",
	}))

	// Create shader module
	shaderCode := `
		@vertex
		fn vs_main(@builtin(vertex_index) vertexIndex: u32) -> @builtin(position) vec4f {
			var pos = array<vec2f, 3>(
				vec2f( 0.0,  0.5),
				vec2f(-0.5, -0.5),
				vec2f( 0.5, -0.5)
			);
			return vec4f(pos[vertexIndex], 0.0, 1.0);
		}

		@fragment
		fn fs_main() -> @location(0) vec4f {
			return vec4f(1.0, 0.0, 0.0, 1.0);
		}
	`

	shaderModule := device.Call("createShaderModule", js.ValueOf(map[string]interface{}{
		"code": shaderCode,
	}))

	// Create render pipeline
	renderPipeline := device.Call("createRenderPipeline", js.ValueOf(map[string]interface{}{
		"layout": "auto",
		"vertex": map[string]interface{}{
			"module":     shaderModule,
			"entryPoint": "vs_main",
		},
		"fragment": map[string]interface{}{
			"module":     shaderModule,
			"entryPoint": "fs_main",
			"targets": []interface{}{
				map[string]interface{}{
					"format": canvasFormat,
				},
			},
		},
		"primitive": map[string]interface{}{
			"topology": "triangle-list",
		},
	}))

	w.pipeline = renderPipeline

	// Create render function
	w.renderFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Get current texture
		texture := w.context.Call("getCurrentTexture")
		if texture.IsUndefined() {
			return nil
		}

		// Create command encoder
		commandEncoder := w.device.Call("createCommandEncoder")

		// Create render pass
		renderPass := commandEncoder.Call("beginRenderPass", js.ValueOf(map[string]interface{}{
			"colorAttachments": []interface{}{
				map[string]interface{}{
					"view":       texture.Call("createView"),
					"clearValue": map[string]interface{}{"r": 0.0, "g": 0.0, "b": 0.0, "a": 1.0},
					"loadOp":     "clear",
					"storeOp":    "store",
				},
			},
		}))

		// Set pipeline and draw
		renderPass.Call("setPipeline", w.pipeline)
		renderPass.Call("draw", 3, 1, 0, 0)
		renderPass.Call("end")

		// Submit command buffer
		commandBuffer := commandEncoder.Call("finish")
		w.device.Get("queue").Call("submit", []interface{}{commandBuffer})

		return nil
	})

	// Initial render
	w.renderFunc.Invoke()

	// Set up animation loop
	var animationLoop js.Func
	animationLoop = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		w.renderFunc.Invoke()
		js.Global().Call("requestAnimationFrame", animationLoop)
		return nil
	})

	// Start the animation loop
	js.Global().Call("requestAnimationFrame", animationLoop)

	println("DEBUG: WebGPU triangle setup complete")
	w.SetStatus(true, "WebGPU triangle rendered successfully!")
}

// initializeWebGL sets up WebGL fallback
func (w *WebGPUCanvasManager) initializeWebGL() {
	println("DEBUG: Setting up WebGL triangle")

	// Get WebGL context
	context := w.canvas.Call("getContext", "webgl")
	if context.IsUndefined() {
		println("DEBUG: WebGL context creation failed")
		w.SetStatus(false, "WebGL not supported in this browser")
		return
	}

	// Set canvas size
	w.canvas.Set("width", 800)
	w.canvas.Set("height", 600)

	// Vertex shader source
	vertexShaderSource := `
		attribute vec2 position;
		void main() {
			gl_Position = vec4(position, 0.0, 1.0);
		}
	`

	// Fragment shader source
	fragmentShaderSource := `
		precision mediump float;
		void main() {
			gl_FragColor = vec4(1.0, 0.0, 0.0, 1.0);
		}
	`

	// Create vertex shader
	vertexShader := context.Call("createShader", 35633) // VERTEX_SHADER
	context.Call("shaderSource", vertexShader, vertexShaderSource)
	context.Call("compileShader", vertexShader)

	// Create fragment shader
	fragmentShader := context.Call("createShader", 35632) // FRAGMENT_SHADER
	context.Call("shaderSource", fragmentShader, fragmentShaderSource)
	context.Call("compileShader", fragmentShader)

	// Create shader program
	program := context.Call("createProgram")
	context.Call("attachShader", program, vertexShader)
	context.Call("attachShader", program, fragmentShader)
	context.Call("linkProgram", program)

	// Use the program
	context.Call("useProgram", program)

	// Create vertex buffer
	vertices := []float32{
		0.0, 0.5, // Top vertex
		-0.5, -0.5, // Bottom left
		0.5, -0.5, // Bottom right
	}

	vertexBuffer := context.Call("createBuffer")
	context.Call("bindBuffer", 34962, vertexBuffer)                // ARRAY_BUFFER
	context.Call("bufferData", 34962, js.ValueOf(vertices), 35044) // STATIC_DRAW

	// Get attribute location
	positionLocation := context.Call("getAttribLocation", program, "position")
	context.Call("enableVertexAttribArray", positionLocation)
	context.Call("vertexAttribPointer", positionLocation, 2, 5126, false, 0, 0) // FLOAT

	// Clear canvas
	context.Call("clearColor", 0.0, 0.0, 0.0, 1.0)
	context.Call("clear", 16384) // COLOR_BUFFER_BIT

	// Draw triangle
	context.Call("drawArrays", 4, 0, 3) // TRIANGLES

	println("DEBUG: WebGL triangle rendered")
	w.SetStatus(true, "WebGL triangle rendered successfully!")
}

// CanvasError represents a canvas-related error
type CanvasError struct {
	Message string
}

func (e *CanvasError) Error() string {
	return e.Message
}

// Helper functions
func getStatusType(initialized bool, message string) string {
	if initialized {
		return "success"
	}
	if message != "" {
		return "warning"
	}
	return "info"
}
