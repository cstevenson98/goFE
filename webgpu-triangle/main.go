package main

import (
	"syscall/js"
)

// Global canvas manager
var canvasManager CanvasManager

func main() {
	println("DEBUG: Go WASM program started")

	// Initialize canvas manager
	canvasManager = NewWebGPUCanvasManager()

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
	}
}
