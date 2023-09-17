package main

import (
	"github.com/cstevenson98/goFE/examples/pokedex/components/entry"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	// Instantiate a new Document
	goFE.Init()
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		entry.NewEntry(entry.Props{}),
	}))
	goFE.GetDocument().Init()
	<-make(chan bool)
}
