package goFE

import (
	"github.com/google/uuid"
	"syscall/js"
)

const renderNotifierBufferSize = 100

type Document struct {
	componentTree []Component

	// When any component's state changes, we should re-render the DOM
	// from this element down
	renderNotifier chan Component
}

// global document
var document *Document

func Init() {
	// Listen for any re-render events
	go func() {
		for {
			select {
			case component := <-document.renderNotifier:
				println("Re-rendering DOM from component with id: " + component.GetID().String())
				rootElement := js.Global().Get("document").Call("getElementById", component.GetID().String())
				rootElement.Set("outerHTML", component.Render())
				initListeners([]Component{component})
			}
		}
	}()
}

func SetDocument(doc *Document) {
	document = doc
}

func GetDocument() *Document {
	return document
}

func NewDocument(componentTree []Component) *Document {
	return &Document{
		componentTree:  componentTree,
		renderNotifier: make(chan Component, renderNotifierBufferSize),
	}
}

func (d *Document) Init() {
	var buffer string

	for _, component := range d.componentTree {
		buffer += component.Render()
	}

	rootElement := js.Global().Get("document").Call("getElementById", "root")
	rootElement.Set("innerHTML", buffer)
	initListeners(d.componentTree)
}

func (d *Document) GetComponentTree() []Component {
	return d.componentTree
}

func (d *Document) Append(component Component) {
	d.componentTree = append(d.componentTree, component)
}

func (d *Document) NotifyRender(component Component) {
	d.renderNotifier <- component
}

func (d *Document) AddEventListener(id uuid.UUID, event string, callback js.Func) {
	println("Adding event listener for component with id: " + id.String())
	js.Global().Get("document").Call("getElementById", id.String()).Call("addEventListener", event, callback)
}

func initListeners(components []Component) {
	for _, component := range components {
		component.InitEventListeners()
		initListeners(component.GetChildren())
	}
}
