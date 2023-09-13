package main

import (
	"github.com/cstevenson98/goFE/internal/components/app"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"syscall/js"
)

func main() {
	// Instantiate a new Document
	goFE.Init()
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		app.NewApp(app.Props{
			Title: "Hello World",
		}),
	}))
	rootElement := js.Global().Get("document").Call("getElementById", "root")
	rootElement.Set("innerHTML", goFE.GetDocument().Init())
	<-make(chan bool)
}
