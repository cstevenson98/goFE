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

func UpdateStateArray[T Component](input *[]T, newLen int, newT func() T) {
	// Children determined by counter state
	if newLen != len(*input) {
		if newLen > len(*input) {
			// Add counters
			for i := len(*input); i < newLen; i++ {
				t := newT()
				*input = append(*input, t)
			}
		} else {
			// Kill the to-be-removed counters
			for i := newLen; i < len(*input); i++ {
				(*input)[i].GetKill() <- true
			}
			*input = (*input)[:newLen]
		}
	}
}
