//go:build js

package engine

import (
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Engine represents the game engine that manages the canvas and game loop
type Engine struct {
	canvasManager canvas.CanvasManager
	lastTime      float64
	textureLoaded bool
	running       bool
}

// NewEngine creates a new game engine instance
func NewEngine() *Engine {
	return &Engine{
		canvasManager: canvas.NewWebGPUCanvasManager(),
		running:       false,
	}
}

// Initialize sets up the engine with the specified canvas ID
func (e *Engine) Initialize(canvasID string) error {
	println("DEBUG: Engine initializing with canvas:", canvasID)

	err := e.canvasManager.Initialize(canvasID)
	if err != nil {
		println("DEBUG: Engine initialization failed:", err.Error())
		return err
	}

	println("DEBUG: Engine initialized successfully")
	return nil
}

// Start begins the game loop
func (e *Engine) Start() {
	if e.running {
		println("DEBUG: Engine already running")
		return
	}

	e.running = true
	println("DEBUG: Engine starting render loop")

	e.startRenderLoop()
}

// startRenderLoop initializes and starts the animation loop
func (e *Engine) startRenderLoop() {
	var animationLoop js.Func
	animationLoop = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !e.running {
			return nil
		}

		currentTime := js.Global().Get("performance").Call("now").Float() / 1000.0

		if e.lastTime == 0 {
			e.lastTime = currentTime
		}

		deltaTime := currentTime - e.lastTime
		e.lastTime = currentTime

		// Update and render the frame
		e.Update(deltaTime)
		e.Render()

		js.Global().Call("requestAnimationFrame", animationLoop)
		return nil
	})

	js.Global().Call("requestAnimationFrame", animationLoop)
}

// Update handles game logic updates
func (e *Engine) Update(deltaTime float64) {
	// Game logic updates go here
	// For now, just handle texture loading
	e.testLlamaTexture()
}

// Render draws the current frame
func (e *Engine) Render() {
	e.canvasManager.Render()
}

// testLlamaTexture loads and draws the test texture
func (e *Engine) testLlamaTexture() {
	// Try to load texture if not loaded yet
	if !e.textureLoaded {
		err := e.canvasManager.LoadTexture("llama.png")
		if err != nil {
			// Canvas not initialized yet, will retry next frame
			return
		}
		e.textureLoaded = true
		println("DEBUG: Texture loading initiated")
		// Give it a moment to load asynchronously
		return
	}

	// Draw llama.png texture at position (100, 100) with size 256x256
	// Full texture UV coordinates
	position := types.Vector2{X: 100, Y: 100}
	size := types.Vector2{X: 256, Y: 256}
	uv := types.UVRect{U: 0, V: 0, W: 1, H: 1}

	err := e.canvasManager.DrawTexturedRect("llama.png", position, size, uv)
	if err != nil {
		// If texture not ready yet, just skip (it's loading asynchronously)
		return
	}
}

// Stop stops the game loop
func (e *Engine) Stop() {
	e.running = false
	println("DEBUG: Engine stopped")
}

// Cleanup releases engine resources
func (e *Engine) Cleanup() error {
	e.Stop()
	return e.canvasManager.Cleanup()
}

// GetCanvasManager returns the underlying canvas manager for advanced usage
func (e *Engine) GetCanvasManager() canvas.CanvasManager {
	return e.canvasManager
}
