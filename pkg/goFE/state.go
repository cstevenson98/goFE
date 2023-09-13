package goFE

type State[T any] struct {
	Value *T
	Chan  chan *T
}

func NewState[T any](value *T) (*State[T], func(*T)) {
	newState := &State[T]{
		Value: value,
		Chan:  make(chan *T),
	}
	setState := func(newValue *T) {
		newState.Chan <- newValue
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
		case value := <-state.Chan:
			state.Value = value
			GetDocument().NotifyRender(component)
		}
	}
}
