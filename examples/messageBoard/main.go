package main

import (
	"github.com/cstevenson98/goFE/examples/messageBoard/messageBoard"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	// Initialize the framework with debug logging
	goFE.Init(&goFE.Logger{
		Level: goFE.DEBUG,
	})

	// Create and set up the message board component
	messageBoardComponent := messageBoard.NewMessageBoard(messageBoard.Props{})

	// Set up the document with the message board as the root component
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		messageBoardComponent,
	}))

	// Initialize the document
	goFE.GetDocument().Init()

	// Keep the program running
	<-make(chan bool)
}
