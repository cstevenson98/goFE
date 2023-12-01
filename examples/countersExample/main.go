package main

import (
	"github.com/cstevenson98/goFE/examples/countersExample/components/counterStack"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	// Instantiate a new Document
	goFE.Init(&goFE.Logger{
		Level: goFE.DEBUG,
	})
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		counterStack.NewCounterStack(counterStack.Props{
			Title: "A selection of counters:",
		}),
	}))
	goFE.GetDocument().Init()
	<-make(chan bool)
}
