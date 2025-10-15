package main

import (
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/engine"
)

// Global engine instance
var gameEngine *engine.Engine

func main() {
	println("DEBUG: Go WASM program started")

	// Create the game engine
	gameEngine = engine.NewEngine()

	// Check if DOM is already loaded
	document := js.Global().Get("document")
	if document.Get("readyState").String() == "complete" {
		println("DEBUG: DOM already loaded, initializing immediately")
		initializeEngine()
	} else {
		println("DEBUG: Waiting for DOM to load")
		// Wait for DOM to be ready
		js.Global().Call("addEventListener", "DOMContentLoaded", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			println("DEBUG: DOMContentLoaded event fired")
			initializeEngine()
			return nil
		}))
	}

	// Keep the program running
	<-make(chan bool)
}

func initializeEngine() {
	println("DEBUG: Starting engine initialization")

	err := gameEngine.Initialize("webgpu-canvas")
	if err != nil {
		println("DEBUG: Engine initialization failed:", err.Error())
		return
	}

	// Start the game loop
	gameEngine.Start()
}
