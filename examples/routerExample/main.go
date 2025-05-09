package main

import (
	"github.com/cstevenson98/goFE/examples/routerExample/components/router"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	println("RouterExample: Starting application")
	
	// Initialize the framework with debug logging
	println("RouterExample: Initializing goFE framework with DEBUG logging")
	goFE.Init(&goFE.Logger{
		Level: goFE.DEBUG,
	})

	// Create and set up the router with route definitions
	println("RouterExample: Creating router component")
	routerComponent := router.NewRouter(router.Props{})

	// Set up the document with the router as the root component
	println("RouterExample: Setting up document with router as root component")
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		routerComponent,
	}))

	// Initialize the document
	println("RouterExample: Initializing document")
	goFE.GetDocument().Init()

	println("RouterExample: Application started and ready")
	
	// Keep the program running
	<-make(chan bool)
} 