package main

import (
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Global variables
var (
	canvasManager canvas.CanvasManager
	lastTime      float64
	textureLoaded bool
)

func main() {
	println("DEBUG: Go WASM program started")

	// Initialize canvas manager
	canvasManager = canvas.NewWebGPUCanvasManager()

	// Check if DOM is already loaded
	document := js.Global().Get("document")
	if document.Get("readyState").String() == "complete" {
		println("DEBUG: DOM already loaded, initializing immediately")
		initializeCanvas()
	} else {
		println("DEBUG: Waiting for DOM to load")
		// Wait for DOM to be ready
		js.Global().Call("addEventListener", "DOMContentLoaded", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			println("DEBUG: DOMContentLoaded event fired")
			initializeCanvas()
			return nil
		}))
	}

	// Keep the program running
	<-make(chan bool)
}

func initializeCanvas() {
	println("DEBUG: Starting canvas initialization")

	err := canvasManager.Initialize("webgpu-canvas")
	if err != nil {
		println("DEBUG: Canvas initialization failed:", err.Error())
		return
	}

	// Start the basic render loop
	startRenderLoop()
}

func startRenderLoop() {
	println("DEBUG: Starting basic render loop")

	// Start the animation loop
	var animationLoop js.Func
	animationLoop = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		currentTime := js.Global().Get("performance").Call("now").Float() / 1000.0

		if lastTime == 0 {
			lastTime = currentTime
		}

		deltaTime := currentTime - lastTime
		lastTime = currentTime

		// Simple render loop
		renderFrame(deltaTime)

		js.Global().Call("requestAnimationFrame", animationLoop)
		return nil
	})

	js.Global().Call("requestAnimationFrame", animationLoop)
}

func renderFrame(deltaTime float64) {
	// Test llama texture
	testLlamaTexture()

	// Render the frame (triangle + sprites)
	canvasManager.Render()
}

func testLlamaTexture() {
	// Try to load texture if not loaded yet
	if !textureLoaded {
		err := canvasManager.LoadTexture("llama.png")
		if err != nil {
			// Canvas not initialized yet, will retry next frame
			return
		}
		textureLoaded = true
		println("DEBUG: Texture loading initiated")
		// Give it a moment to load asynchronously
		return
	}

	// Draw llama.png texture at position (100, 100) with size 256x256
	// Full texture UV coordinates
	position := types.Vector2{X: 100, Y: 100}
	size := types.Vector2{X: 256, Y: 256}
	uv := types.UVRect{U: 0, V: 0, W: 1, H: 1}

	err := canvasManager.DrawTexturedRect("llama.png", position, size, uv)
	if err != nil {
		// If texture not ready yet, just skip (it's loading asynchronously)
		return
	}
}

func testMultipleRectangles() {
	// Start batch mode
	err := canvasManager.BeginBatch()
	if err != nil {
		println("DEBUG: Failed to begin batch:", err.Error())
		return
	}

	// Draw 5 rectangles with different colors and positions
	rectangles := []struct {
		pos   types.Vector2
		size  types.Vector2
		color [4]float32
		name  string
	}{
		{types.Vector2{X: 100, Y: 100}, types.Vector2{X: 64, Y: 64}, [4]float32{1.0, 0.0, 0.0, 1.0}, "Red"},     // Red - top left
		{types.Vector2{X: 200, Y: 100}, types.Vector2{X: 64, Y: 64}, [4]float32{0.0, 1.0, 0.0, 1.0}, "Green"},   // Green - top center-left
		{types.Vector2{X: 300, Y: 100}, types.Vector2{X: 64, Y: 64}, [4]float32{0.0, 0.5, 1.0, 1.0}, "Blue"},    // Blue - top center
		{types.Vector2{X: 400, Y: 100}, types.Vector2{X: 64, Y: 64}, [4]float32{1.0, 1.0, 0.0, 1.0}, "Yellow"},  // Yellow - top center-right
		{types.Vector2{X: 500, Y: 100}, types.Vector2{X: 64, Y: 64}, [4]float32{1.0, 0.0, 1.0, 1.0}, "Magenta"}, // Magenta - top right
	}

	for _, rect := range rectangles {
		err = canvasManager.DrawColoredRect(rect.pos, rect.size, rect.color)
		if err != nil {
			println("DEBUG: Failed to draw", rect.name, "rectangle:", err.Error())
		}
	}

	// End batch mode (uploads all vertices at once)
	err = canvasManager.EndBatch()
	if err != nil {
		println("DEBUG: Failed to end batch:", err.Error())
	}
}
