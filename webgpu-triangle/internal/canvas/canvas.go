package canvas

import (
	"syscall/js"

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

	// New fields for sprite rendering
	spritePipeline     js.Value
	backgroundPipeline js.Value
	vertexBuffer       js.Value
	uniformBuffer      js.Value
	bindGroup          js.Value
	sampler            js.Value
	batchMode          bool
	vertices           []types.SpriteVertex

	// Staged sprite vertices for rendering
	stagedVertices    []float32
	stagedVertexCount int
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
	if !w.initialized {
		return nil // Silently skip if not initialized yet
	}

	w.renderFrame()
	return nil
}

// renderFrame performs the actual rendering (called by Render or animation loop)
func (w *WebGPUCanvasManager) renderFrame() {
	if w.context.IsUndefined() || w.device.IsUndefined() {
		return
	}

	// Get current texture
	texture := w.context.Call("getCurrentTexture")
	if texture.IsUndefined() {
		return
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

	// Draw triangle
	renderPass.Call("setPipeline", w.pipeline)
	renderPass.Call("draw", 3, 1, 0, 0)

	// Draw sprites if we have any staged vertices
	if w.stagedVertexCount > 0 && !w.spritePipeline.IsUndefined() {
		renderPass.Call("setPipeline", w.spritePipeline)
		renderPass.Call("setVertexBuffer", 0, w.vertexBuffer)
		renderPass.Call("draw", w.stagedVertexCount, 1, 0, 0)
	}

	renderPass.Call("end")

	// Submit command buffer
	commandBuffer := commandEncoder.Call("finish")
	w.device.Get("queue").Call("submit", []interface{}{commandBuffer})

	// Clear staged vertices after rendering
	w.stagedVertexCount = 0
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

	// Create sprite pipeline for rectangle/texture rendering
	println("DEBUG: Creating sprite pipeline")
	w.createSpritePipeline(device, canvasFormat)

	// Create vertex buffer for sprites
	println("DEBUG: Creating sprite vertex buffer")
	w.createSpriteVertexBuffer(device)

	// Create render function - no longer auto-loops
	w.renderFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		w.renderFrame()
		return nil
	})

	// Don't start auto animation loop - let main.go control it
	println("DEBUG: WebGPU render function created (controlled by main.go)")

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

// New rendering methods implementation

// DrawTexture draws a texture at the specified position and size
func (w *WebGPUCanvasManager) DrawTexture(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	// For Phase 2, support custom colors (later will support textures)
	// For now, use blue as default
	color := [4]float32{0.0, 0.5, 1.0, 1.0} // Blue color (RGBA)

	return w.DrawColoredRect(position, size, color)
}

// DrawColoredRect draws a colored rectangle with the specified color
func (w *WebGPUCanvasManager) DrawColoredRect(position types.Vector2, size types.Vector2, color [4]float32) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}
	return w.drawColoredRect(position, size, color)
}

// drawColoredRect draws a colored rectangle (internal helper)
func (w *WebGPUCanvasManager) drawColoredRect(position types.Vector2, size types.Vector2, color [4]float32) error {
	// Generate vertices for the rectangle
	vertices := w.generateQuadVertices(position, size, color)

	if w.batchMode {
		// In batch mode, accumulate vertices
		w.stagedVertices = append(w.stagedVertices, vertices...)
		println("DEBUG: Batched rectangle at", position.X, position.Y, "- Total vertices:", len(w.stagedVertices))
	} else {
		// Immediate mode - upload and stage for rendering
		verticesTypedArray := js.Global().Get("Float32Array").New(len(vertices))
		for i, v := range vertices {
			verticesTypedArray.SetIndex(i, v)
		}

		w.device.Get("queue").Call("writeBuffer",
			w.vertexBuffer,
			0, // offset
			verticesTypedArray,
			0,             // data offset
			len(vertices), // size in floats
		)

		// Stage the vertex count for rendering (6 vertices = 1 quad)
		w.stagedVertexCount = len(vertices) / 6 // 6 floats per vertex

		println("DEBUG: Immediate mode - Staged", w.stagedVertexCount, "vertices")
	}

	return nil
}

// DrawTextureRotated draws a rotated texture (STUB)
func (w *WebGPUCanvasManager) DrawTextureRotated(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, rotation float64) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	println("DEBUG: DrawTextureRotated STUB - Position:", position.X, position.Y, "Rotation:", rotation)
	// TODO: Implement actual rotated texture drawing
	return nil
}

// DrawTextureScaled draws a scaled texture
func (w *WebGPUCanvasManager) DrawTextureScaled(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, scale types.Vector2) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	println("DEBUG: DrawTextureScaled STUB - Position:", position.X, position.Y, "Scale:", scale.X, scale.Y)
	// TODO: Implement actual scaled texture drawing

	return nil
}

// BeginBatch starts batch rendering mode
func (w *WebGPUCanvasManager) BeginBatch() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	w.batchMode = true
	w.stagedVertices = make([]float32, 0) // Clear any previous staged vertices
	w.stagedVertexCount = 0

	println("DEBUG: BeginBatch - Batch mode enabled")
	return nil
}

// EndBatch ends batch rendering mode and uploads all batched vertices
func (w *WebGPUCanvasManager) EndBatch() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	if !w.batchMode {
		println("DEBUG: EndBatch called but not in batch mode")
		return nil
	}

	// Upload all accumulated vertices to GPU
	err := w.FlushBatch()
	if err != nil {
		return err
	}

	w.batchMode = false
	println("DEBUG: EndBatch - Batch mode disabled,", w.stagedVertexCount, "vertices uploaded")

	return nil
}

// FlushBatch uploads accumulated vertices to GPU without leaving batch mode
func (w *WebGPUCanvasManager) FlushBatch() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	if len(w.stagedVertices) == 0 {
		println("DEBUG: FlushBatch - No vertices to flush")
		w.stagedVertexCount = 0
		return nil
	}

	// Upload all vertices to GPU
	verticesTypedArray := js.Global().Get("Float32Array").New(len(w.stagedVertices))
	for i, v := range w.stagedVertices {
		verticesTypedArray.SetIndex(i, v)
	}

	w.device.Get("queue").Call("writeBuffer",
		w.vertexBuffer,
		0, // offset
		verticesTypedArray,
		0,                     // data offset
		len(w.stagedVertices), // size in floats
	)

	// Calculate vertex count (6 floats per vertex)
	w.stagedVertexCount = len(w.stagedVertices) / 6

	println("DEBUG: FlushBatch - Uploaded", len(w.stagedVertices), "floats (", w.stagedVertexCount, "vertices )")

	return nil
}

// GetSpritePipeline returns the sprite rendering pipeline (STUB)
func (w *WebGPUCanvasManager) GetSpritePipeline() types.Pipeline {
	println("DEBUG: GetSpritePipeline STUB")
	// TODO: Implement sprite pipeline
	return &types.WebGPUPipeline{Valid: false}
}

// GetBackgroundPipeline returns the background rendering pipeline (STUB)
func (w *WebGPUCanvasManager) GetBackgroundPipeline() types.Pipeline {
	println("DEBUG: GetBackgroundPipeline STUB")
	// TODO: Implement background pipeline
	return &types.WebGPUPipeline{Valid: false}
}

// createSpritePipeline creates the rendering pipeline for sprites
func (w *WebGPUCanvasManager) createSpritePipeline(device js.Value, canvasFormat js.Value) error {
	println("DEBUG: Creating sprite shaders")

	// Sprite shader code (colored rectangles - no textures yet)
	spriteShaderCode := `
		struct VertexOutput {
			@builtin(position) position: vec4f,
			@location(0) color: vec4f,
		}

		@vertex
		fn vs_main(
			@location(0) position: vec2f,
			@location(1) color: vec4f
		) -> VertexOutput {
			var output: VertexOutput;
			output.position = vec4f(position, 0.0, 1.0);
			output.color = color;
			return output;
		}

		@fragment
		fn fs_main(@location(0) color: vec4f) -> @location(0) vec4f {
			return color;
		}
	`

	spriteShaderModule := device.Call("createShaderModule", js.ValueOf(map[string]interface{}{
		"code": spriteShaderCode,
	}))

	println("DEBUG: Creating sprite pipeline")

	// Create sprite render pipeline with vertex buffer layout
	spritePipeline := device.Call("createRenderPipeline", js.ValueOf(map[string]interface{}{
		"layout": "auto",
		"vertex": map[string]interface{}{
			"module":     spriteShaderModule,
			"entryPoint": "vs_main",
			"buffers": []interface{}{
				map[string]interface{}{
					"arrayStride": 24, // 6 floats * 4 bytes = 24 bytes (x, y, r, g, b, a)
					"attributes": []interface{}{
						map[string]interface{}{
							"shaderLocation": 0,
							"offset":         0,
							"format":         "float32x2", // position (x, y)
						},
						map[string]interface{}{
							"shaderLocation": 1,
							"offset":         8,
							"format":         "float32x4", // color (r, g, b, a)
						},
					},
				},
			},
		},
		"fragment": map[string]interface{}{
			"module":     spriteShaderModule,
			"entryPoint": "fs_main",
			"targets": []interface{}{
				map[string]interface{}{
					"format": canvasFormat,
					"blend": map[string]interface{}{
						"color": map[string]interface{}{
							"srcFactor": "src-alpha",
							"dstFactor": "one-minus-src-alpha",
							"operation": "add",
						},
						"alpha": map[string]interface{}{
							"srcFactor": "one",
							"dstFactor": "one-minus-src-alpha",
							"operation": "add",
						},
					},
				},
			},
		},
		"primitive": map[string]interface{}{
			"topology": "triangle-list",
		},
	}))

	w.spritePipeline = spritePipeline
	println("DEBUG: Sprite pipeline created successfully")

	return nil
}

// createSpriteVertexBuffer creates a dynamic vertex buffer for sprite rendering
func (w *WebGPUCanvasManager) createSpriteVertexBuffer(device js.Value) error {
	println("DEBUG: Creating sprite vertex buffer")

	// Create a buffer large enough for multiple sprites (start with 1024 vertices)
	bufferSize := 1024 * 24 // 1024 vertices * 24 bytes per vertex

	vertexBuffer := device.Call("createBuffer", js.ValueOf(map[string]interface{}{
		"size":  bufferSize,
		"usage": js.Global().Get("GPUBufferUsage").Get("VERTEX").Int() | js.Global().Get("GPUBufferUsage").Get("COPY_DST").Int(),
	}))

	w.vertexBuffer = vertexBuffer
	w.vertices = make([]types.SpriteVertex, 0)

	println("DEBUG: Sprite vertex buffer created, size:", bufferSize)

	return nil
}

// canvasToNDC converts canvas coordinates to Normalized Device Coordinates
func (w *WebGPUCanvasManager) canvasToNDC(x, y float64) (float32, float32) {
	width := w.canvas.Get("width").Float()
	height := w.canvas.Get("height").Float()

	// Convert to NDC (-1 to 1)
	// X: left = -1, right = 1
	// Y: top = 1, bottom = -1 (inverted from canvas where top = 0)
	ndcX := float32((x/width)*2.0 - 1.0)
	ndcY := float32(1.0 - (y/height)*2.0) // Flip Y axis

	return ndcX, ndcY
}

// generateQuadVertices generates vertices for a colored rectangle
func (w *WebGPUCanvasManager) generateQuadVertices(pos types.Vector2, size types.Vector2, color [4]float32) []float32 {
	// Calculate corners in canvas coordinates
	x0 := pos.X
	y0 := pos.Y
	x1 := pos.X + size.X
	y1 := pos.Y + size.Y

	// Convert to NDC
	ndcX0, ndcY0 := w.canvasToNDC(x0, y0)
	ndcX1, ndcY1 := w.canvasToNDC(x1, y1)

	// Generate 6 vertices for 2 triangles (quad)
	// Each vertex: [x, y, r, g, b, a]
	vertices := []float32{
		// Triangle 1
		ndcX0, ndcY0, color[0], color[1], color[2], color[3], // Top-left
		ndcX1, ndcY0, color[0], color[1], color[2], color[3], // Top-right
		ndcX0, ndcY1, color[0], color[1], color[2], color[3], // Bottom-left

		// Triangle 2
		ndcX1, ndcY0, color[0], color[1], color[2], color[3], // Top-right
		ndcX1, ndcY1, color[0], color[1], color[2], color[3], // Bottom-right
		ndcX0, ndcY1, color[0], color[1], color[2], color[3], // Bottom-left
	}

	return vertices
}

// ClearCanvas clears the canvas
func (w *WebGPUCanvasManager) ClearCanvas() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	canvas := w.canvas
	if canvas.IsUndefined() || canvas.IsNull() {
		println("DEBUG: Canvas is undefined or null, cannot clear")
		return nil
	}

	ctx := canvas.Call("getContext", "2d")
	if ctx.IsUndefined() || ctx.IsNull() {
		println("DEBUG: 2D context is undefined or null, cannot clear")
		return nil
	}

	// Clear the entire canvas
	width := canvas.Get("width")
	height := canvas.Get("height")

	if !width.IsUndefined() && !height.IsUndefined() {
		ctx.Call("clearRect", 0, 0, width.Float(), height.Float())
		println("DEBUG: Canvas cleared")
	} else {
		println("DEBUG: Canvas dimensions not available")
	}

	return nil
}
