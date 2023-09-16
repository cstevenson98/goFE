package main

import (
	"github.com/cstevenson98/goFE/internal/components/counterStack"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	// Instantiate a new Document
	goFE.Init()
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		counterStack.NewApp(counterStack.Props{
			Title: "A selection of counters:",
		}),
	}))

	goFE.GetDocument().Init()
	println("Initialized document", goFE.GetDocument().GetComponentTree())
	<-make(chan bool)
}
