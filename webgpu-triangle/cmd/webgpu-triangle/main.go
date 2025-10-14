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
	// Stage sprites for rendering
	testBasicRectangle()

	// Render the frame (triangle + sprites)
	canvasManager.Render()
}

func testBasicRectangle() {
	// Create a simple texture stub (not actually used in Phase 1)
	texture := &types.WebGPUTexture{
		Width:  64,
		Height: 64,
		ID:     "test",
	}

	// Draw a blue rectangle at position (100, 100) with size 64x64
	position := types.Vector2{X: 100, Y: 100}
	size := types.Vector2{X: 64, Y: 64}
	uv := types.UVRect{U: 0, V: 0, W: 1, H: 1}

	err := canvasManager.DrawTexture(texture, position, size, uv)
	if err != nil {
		println("DEBUG: Failed to draw texture:", err.Error())
	}
}
