package goFE

import (
	"github.com/google/uuid"
	"sync"
)

// State is a generic struct that holds a value and a channel
type State[T any] struct {
	Value     *T
	id        uuid.UUID
	ch        chan *T
	lock      sync.Mutex
	listeners map[uuid.UUID]func(value *T)
}

// NewState creates a new instance of frontend state. It returns a pointer to the
// new state, with initial value, and a function to set the state.
func NewState[T any](component Component, value *T) (*State[T], func(*T)) {
	newState := &State[T]{
		Value:     value,
		id:        uuid.New(),
		ch:        make(chan *T),
		listeners: make(map[uuid.UUID]func(value *T)),
	}
	setState := func(newValue *T) {
		newState.ch <- newValue
	}
	component.AddKill(newState.id)
	go listenForStateChange[T](component, newState)
	return newState, setState
}

// AddEffect adds an effect to the state. An effect is a function that is called
// whenever the state changes.
func (s *State[T]) AddEffect(effect func(value *T)) {
	s.lock.Lock()
	s.listeners[uuid.New()] = effect
	s.lock.Unlock()
}

// listenForStateChange listens for state changes and updates the state accordingly.
func listenForStateChange[T any](component Component, state *State[T]) {
	//println("Listening for state change, componentID: ", component.GetID().String())
	for {
		select {
		case value := <-state.ch:
			state.lock.Lock()
			//println("State change detected, componentID: ", component.GetID().String())
			state.Value = value
			state.lock.Unlock()
			document.renderNotifier <- component
			notifyListeners[T](state)
		case <-component.GetKill(state.id):
			//println("Stopped listening for state change, componentID: ", component.GetID().String())
			// kill child components
			for _, child := range component.GetChildren() {
				child.KillAll()
			}
			return
		}
	}
}

// notifyListeners notifies all listeners of a state change.
func notifyListeners[T any](state *State[T]) {
	state.lock.Lock()
	for _, listener := range state.listeners {
		go listener(state.Value)
	}
	state.lock.Unlock()
}

// UpdateComponentArray provides functionality to control a variable-length collection of components,
// such as a list of rows in a table, or any other collection of sub-components (children).
func UpdateComponentArray[T Component, Props any](input *[]T, newLen int, newT func(props *Props) T, newProps []*Props) {
	if input == nil {
		panic("'UpdateComponentArray' input cannot be nil")
	}
	if newProps != nil {
		// kill all components and rebuild
		for _, component := range *input {
			component.KillAll()
		}
		*input = nil
		for i := 0; i < newLen; i++ {
			var t T
			t = newT(newProps[i])
			*input = append(*input, t)
		}
		return
	}
	if newLen != len(*input) {
		if newLen > len(*input) {
			// Add components
			for i := len(*input); i < newLen; i++ {
				var t T
				t = newT(nil)
				*input = append(*input, t)
			}
		} else {
			// GetKill the to-be-removed components
			for i := newLen; i < len(*input); i++ {
				(*input)[i].KillAll()
			}
			*input = (*input)[:newLen]
		}
	}
}
