package main

import (
	"github.com/cstevenson98/goFE/internal/components/app"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	// Instantiate a new Document
	goFE.Init()
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		app.NewApp(app.Props{
			Title: "Hello World",
		}),
	}))
	goFE.GetDocument().Init()
	println("Initialized document", goFE.GetDocument().GetComponentTree())
	<-make(chan bool)
}
