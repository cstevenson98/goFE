package main

import (
	"github.com/cstevenson98/goFE/examples/anthropicAgentExample/components"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	println("AnthropicAgentExample: Starting application")

	// Initialize the framework with debug logging
	println("AnthropicAgentExample: Initializing goFE framework with DEBUG logging")
	goFE.Init(&goFE.Logger{
		Level: goFE.DEBUG,
	})

	// Create and set up the Anthropic Agent example component
	println("AnthropicAgentExample: Creating Anthropic Agent example component")
	anthropicAgentComponent := components.NewAnthropicAgentExample()

	// Set up the document with the Anthropic Agent example as the root component
	println("AnthropicAgentExample: Setting up document with Anthropic Agent example as root component")
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		anthropicAgentComponent,
	}))

	// Initialize the document
	println("AnthropicAgentExample: Initializing document")
	goFE.GetDocument().Init()

	println("AnthropicAgentExample: Application started and ready")

	// Keep the program running
	<-make(chan bool)
}
