package main

import (
	"github.com/cstevenson98/goFE/examples/pokedex/pokedex"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	// Instantiate a new Document
	goFE.Init(&goFE.Logger{
		Level: goFE.DEBUG,
	})
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		pokedex.NewPokedex(pokedex.Props{}),
	}))
	goFE.GetDocument().Init()
	<-make(chan bool)
}
