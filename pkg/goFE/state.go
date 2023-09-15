package goFE

import "sync"

// State is a generic struct that holds a value and a channel
type State[T any] struct {
	Value *T
	ch    chan *T
	lock  sync.Mutex
}

// NewState creates a new instance of frontend state. It returns a pointer to the
// new state, with initial value, and a function to set the state.
func NewState[T any](value *T) (*State[T], func(*T)) {
	newState := &State[T]{
		Value: value,
		ch:    make(chan *T),
	}
	setState := func(newValue *T) {
		newState.ch <- newValue
	}
	return newState, setState
}

func ListenForStateChange[T any](component Component, state *State[T]) {
	println("Listening for state change, componentID: ", component.GetID().String())
	for {
		select {
		case <-component.GetKill():
			println("Stopped listening for state change, componentID: ", component.GetID().String())
			return
		case value := <-state.ch:
			state.lock.Lock()
			println("State change detected, componentID: ", component.GetID().String())
			state.Value = value
			state.lock.Unlock()
			GetDocument().NotifyRender(component)
		}
	}
}
