//go:generate go run github.com/valyala/quicktemplate/qtc

package router

import (
	"strings"
	"syscall/js"

	"github.com/cstevenson98/goFE/examples/messageBoard/messageBoard"
	"github.com/cstevenson98/goFE/examples/pokedex/pokedex"
	"github.com/cstevenson98/goFE/examples/routerExample/components/about"
	"github.com/cstevenson98/goFE/examples/routerExample/components/contact"
	"github.com/cstevenson98/goFE/examples/routerExample/components/home"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
)

// Props defines the router props
type Props struct{}

// routerState contains the current route path
type routerState struct {
	currentPath string
}

// View is a function type that creates a component for a route
type ViewCreator func() goFE.Component

// Router manages navigation and rendering of appropriate views
type Router struct {
	id               uuid.UUID
	navContainer     uuid.UUID
	contentArea      uuid.UUID
	state            *goFE.State[routerState]
	setState         func(*routerState)
	routes           map[string]ViewCreator
	currentView      *goFE.SwappableComponent
	popStateFunc     js.Func
	lastRenderedPath string // Track the last path that was rendered
}

// NewRouter creates a new router component
func NewRouter(_ Props) *Router {
	println("Router: Creating new router component")
	r := &Router{
		id:           uuid.New(),
		navContainer: uuid.New(),
		contentArea:  uuid.New(),
		routes: map[string]ViewCreator{
			"/":             func() goFE.Component { return home.NewHome(home.Props{}) },
			"/about":        func() goFE.Component { return about.NewAbout(about.Props{}) },
			"/contact":      func() goFE.Component { return contact.NewContact(contact.Props{}) },
			"/pokedex":      func() goFE.Component { return pokedex.NewPokedex(pokedex.Props{}) },
			"/messageboard": func() goFE.Component { return messageBoard.NewMessageBoard(messageBoard.Props{}) },
		},
	}

	// Get initial path from browser
	initialPath := js.Global().Get("window").Get("location").Get("pathname").String()
	if initialPath == "" {
		initialPath = "/"
	}

	println("Router: Initializing with path:", initialPath)

	// Create the initial view for the path
	initialViewCreator, exists := r.routes[initialPath]
	if !exists {
		initialPath = "/"
		initialViewCreator = r.routes["/"]
	}

	// Create initial component and wrap it in a SwappableComponent
	initialComponent := initialViewCreator()
	r.currentView = goFE.NewSwappableComponent(initialComponent)
	r.lastRenderedPath = initialPath
	println("Router: Initial view created with ID:", initialComponent.GetID().String())

	// Initialize state with path
	r.state, r.setState = goFE.NewState[routerState](r, &routerState{
		currentPath: initialPath,
	})
	println("Router: State initialized with path")

	return r
}

// GetID returns the component ID
func (r *Router) GetID() uuid.UUID {
	return r.id
}

// Render renders the router component
func (r *Router) Render() string {
	println("Router: Render called, current path:", r.state.Value.currentPath)

	// Update the current view if needed
	r.updateCurrentView(r.state.Value.currentPath)

	// Render the content
	viewContent := ""
	if r.currentView != nil && r.currentView.GetCurrent() != nil {
		println("Router: Rendering child view with ID:", r.currentView.GetCurrent().GetID().String())
		viewContent = r.currentView.Render()
	} else {
		println("Router: No current view to render")
	}

	return RouterTemplate(r.id.String(), r.navContainer.String(), r.contentArea.String(), viewContent, r.state.Value.currentPath)
}

// GetChildren returns child components
func (r *Router) GetChildren() []goFE.Component {
	println("Router: GetChildren called")
	if r.currentView == nil {
		println("Router: No children to return")
		return nil
	}
	if current := r.currentView.GetCurrent(); current != nil {
		println("Router: Returning current view as child:", current.GetID().String())
		// SwappableComponent itself implements Component, so we return it
		return []goFE.Component{r.currentView}
	}
	println("Router: No children to return")
	return nil
}

// InitEventListeners sets up event listeners for navigation
func (r *Router) InitEventListeners() {
	println("Router: Initializing event listeners")

	// Setup click listeners for navigation links
	doc := goFE.GetDocument()
	doc.AddEventListener(r.navContainer, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Get the event object
		event := args[0]

		// Get the target element that was clicked
		target := event.Get("target")

		// Check if we clicked on an anchor tag
		if target.Get("tagName").String() == "A" {
			// Prevent default navigation
			event.Call("preventDefault")

			// Get the href attribute safely
			href := target.Call("getAttribute", "href").String()

			println("Router: Navigation link clicked, href:", href)

			// Navigate to the new path
			r.navigateTo(href)
		}
		return nil
	}))

	// Handle browser back/forward buttons with popstate event
	r.popStateFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		path := js.Global().Get("window").Get("location").Get("pathname").String()
		if path == "" {
			path = "/"
		}

		println("Router: Popstate event detected, path:", path)

		// Fully update the view when the popstate event occurs
		r.navigateTo(path)
		return nil
	})

	println("Router: Adding popstate event listener")
	js.Global().Get("window").Call("addEventListener", "popstate", r.popStateFunc)
}

// navigateTo navigates to a specific path
func (r *Router) navigateTo(path string) {
	println("Router: Navigating to path:", path)

	// Skip if we're already on this path
	if r.state.Value.currentPath == path {
		println("Router: Already on path:", path)
		return
	}

	// Update the URL in the browser
	js.Global().Get("window").Get("history").Call("pushState", nil, "", path)

	// Update the component state
	r.setState(&routerState{
		currentPath: path,
	})
}

// updateCurrentView updates the current view based on the path
func (r *Router) updateCurrentView(path string) {
	println("Router: updateCurrentView called for path:", path)

	// Clean up path (remove trailing slash except for root)
	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
		println("Router: Cleaned up path to:", path)
	}

	// Check if we already rendered this path
	if r.lastRenderedPath == path {
		println("Router: View already rendered for path:", path)
		return
	}

	// Find the view creator for this path
	viewCreator, exists := r.routes[path]

	// If route exists, create the view
	if exists {
		println("Router: Route exists for path:", path)

		// Create new component
		newComponent := viewCreator()
		println("Router: Created new component with ID:", newComponent.GetID().String())

		// Swap the component in the SwappableComponent
		// This will automatically clean up the old component
		r.currentView.Swap(newComponent)
		println("Router: Swapped to new component")

		// Update the last rendered path
		r.lastRenderedPath = path
	} else {
		println("Router: Route not found for path:", path, "redirecting to home")
		// Route not found, redirect to home
		if path != "/" {
			r.navigateTo("/")
		} else {
			println("Router: Already on home path, no redirect needed")
		}
	}
}
