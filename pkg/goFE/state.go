package goFE

type State[T any] struct {
	Value T
	Chan  chan *T
}

func ListenForStateChange[T any](state State[T], kill chan bool) {
	for {
		select {
		case <-kill:
			return
		case value := <-state.Chan:
			state.Value = *value
		}
	}
}
