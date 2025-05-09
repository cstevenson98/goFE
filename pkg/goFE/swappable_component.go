package goFE

import (
	"github.com/google/uuid"
)

// SwappableComponent manages a dynamic component that can be swapped at runtime
// It properly cleans up resources (particularly go routines) from the previous component
type SwappableComponent struct {
	id             uuid.UUID
	current        Component
}

// NewSwappableComponent creates a new SwappableComponent with an optional initial component
func NewSwappableComponent(initialComponent Component) *SwappableComponent {
	sc := &SwappableComponent{
		id:      uuid.New(),
		current: initialComponent,
	}
	
	return sc
}

// GetID returns the SwappableComponent's ID
func (sc *SwappableComponent) GetID() uuid.UUID {
	return sc.id
}

// GetCurrent returns the current active component
func (sc *SwappableComponent) GetCurrent() Component {
	return sc.current
}

// Swap replaces the current component with a new one
// It ensures that resources from the old component are properly cleaned up
func (sc *SwappableComponent) Swap(newComponent Component) {
	// Clean up the old component if it exists
	if sc.current != nil {
		// Kill all states associated with the old component and its children
		// This will stop all go routines associated with the state
		killAllStates(sc.current)
		
		// Log the cleanup
		logger.Log(DEBUG, "SwappableComponent: Cleaned up previous component: "+sc.current.GetID().String())
	}
	
	// Set the new component
	sc.current = newComponent
	
	if newComponent != nil {
		logger.Log(DEBUG, "SwappableComponent: Set new component: "+newComponent.GetID().String())
	} else {
		logger.Log(DEBUG, "SwappableComponent: Set component to nil")
	}
}

// Implements Component interface to delegate to the current component

// Render delegates rendering to the current component
func (sc *SwappableComponent) Render() string {
	if sc.current == nil {
		return ""
	}
	return sc.current.Render()
}

// GetChildren delegates to the current component
func (sc *SwappableComponent) GetChildren() []Component {
	if sc.current == nil {
		return nil
	}
	return sc.current.GetChildren()
}

// InitEventListeners delegates to the current component
func (sc *SwappableComponent) InitEventListeners() {
	if sc.current == nil {
		return
	}
	sc.current.InitEventListeners()
}

// Cleanup explicitly cleans up resources and removes the current component
func (sc *SwappableComponent) Cleanup() {
	sc.Swap(nil) // Swap with nil to clean up the current component
} 