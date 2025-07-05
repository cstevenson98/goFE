package main

import (
	"github.com/cstevenson98/goFE/examples/apiExample/components"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	println("APIExample: Starting application")

	// Initialize the framework with debug logging
	println("APIExample: Initializing goFE framework with DEBUG logging")
	goFE.Init(&goFE.Logger{
		Level: goFE.DEBUG,
	})

	// Create and set up the API example component
	println("APIExample: Creating API example component")
	apiExampleComponent := components.NewAPIExample()

	// Set up the document with the API example as the root component
	println("APIExample: Setting up document with API example as root component")
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		apiExampleComponent,
	}))

	// Initialize the document
	println("APIExample: Initializing document")
	goFE.GetDocument().Init()

	println("APIExample: Application started and ready")

	// Keep the program running
	<-make(chan bool)
}
