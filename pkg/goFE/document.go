package goFE

import (
	"github.com/google/uuid"
	"sync"
	"syscall/js"
)

const renderNotifierBufferSize = 100

type Document struct {
	componentTree []Component

	// Whenever a component is created, we should add a kill switch to this map
	killSwitches map[uuid.UUID]chan bool
	ksLock       sync.Mutex

	// When any component's state changes, we should re-render the DOM
	// from this element down
	renderNotifier chan Component

	// The listeners that are active on the document. We should check when a
	// re-render happens that we don't re-add listeners that are already there
	activeListeners map[uuid.UUID]bool
	listenersLock   sync.Mutex
}

// global document
var document *Document
var listenerCount int

func Init() {
	// Listen for any re-render events
	go func() {
		for {
			select {
			case component := <-document.renderNotifier:
				// Re-render the DOM from the component with the given id down
				println("Re-rendering DOM from component with id: " + component.GetID().String())
				rootElement := js.Global().Get("document").Call("getElementById", component.GetID().String())
				rootElement.Set("innerHTML", component.Render())
				initListeners([]Component{component})
				println("Listener count: ", listenerCount)
			}
		}
	}()

	// A test ticker which triggers a re-render on a random uuid
	//go func() {
	//	for {
	//		select {
	//		case <-time.After(1 * time.Second):
	//			document.renderNotifier <- uuid.New()
	//		}
	//	}
	//}()
}

func SetDocument(doc *Document) {
	document = doc
}

func GetDocument() *Document {
	return document
}

func NewDocument(componentTree []Component) *Document {
	return &Document{
		componentTree:   componentTree,
		killSwitches:    make(map[uuid.UUID]chan bool),
		renderNotifier:  make(chan Component, renderNotifierBufferSize),
		activeListeners: make(map[uuid.UUID]bool),
	}
}

func (d *Document) Init() {
	var buffer string

	for _, component := range d.componentTree {
		buffer += component.Render()
	}

	//initKillSwitches(d)
	rootElement := js.Global().Get("document").Call("getElementById", "root")
	rootElement.Set("innerHTML", buffer)
	initListeners(d.componentTree)
}

func initKillSwitches(d *Document) {
	for _, component := range d.componentTree {
		d.killSwitches[component.GetID()] = component.GetKill()
	}
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
	// Check if the listener is already active
	_, ok := d.activeListeners[id]
	if ok {
		return
	}
	// Add the listener
	js.Global().Get("document").Call("getElementById", id.String()).Call("addEventListener", event, callback)
	d.listenersLock.Lock()
	defer d.listenersLock.Unlock()
	d.activeListeners[id] = true
	listenerCount++
}

func initListeners(components []Component) {
	for _, component := range components {
		component.InitEventListeners()
		initListeners(component.GetChildren())
	}
}
