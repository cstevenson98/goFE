package goFE

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

var stateLock sync.Mutex
var stateKillChannels map[uuid.UUID]map[uuid.UUID]chan bool

func init() {
	stateKillChannels = make(map[uuid.UUID]map[uuid.UUID]chan bool)
	go func() {
		for {
			total := 0
			for _, componentMap := range stateKillChannels {
				total += len(componentMap)
			}
			println("Total number of kill channels", total)
			<-time.After(5 * time.Second)
		}
	}()
}

func registerComponentIfNotExists(component Component) {
	println("Registering component if not exists, componentID: ", component.GetID().String())
	if _, ok := stateKillChannels[component.GetID()]; !ok {
		stateKillChannels[component.GetID()] = make(map[uuid.UUID]chan bool)
	}
	println("Registered component: ", component.GetID().String())
}

func registerKillChannel[T any](component Component, state *State[T]) {
	logger.Log(DEBUG, "Registering kill channel, componentID: "+component.GetID().String())
	stateLock.Lock()
	defer stateLock.Unlock()
	registerComponentIfNotExists(component)
	ch := make(chan bool)
	stateKillChannels[component.GetID()][state.id] = ch
	state.kill = ch
}

func killAllStates(component Component) {
	logger.Log(DEBUG, "Killing all states, componentID: "+component.GetID().String())
	stateLock.Lock()
	defer stateLock.Unlock()
	if _, ok := stateKillChannels[component.GetID()]; ok {
		for _, killCh := range stateKillChannels[component.GetID()] {
			killCh <- true
		}
		delete(stateKillChannels, component.GetID())
	} else {
		logger.Log(WARNING, "No states to kill, componentID: "+component.GetID().String())
	}
}

// State is a generic struct that holds a value and a channel
type State[T any] struct {
	Value     *T
	id        uuid.UUID
	ch        chan *T
	lock      sync.Mutex
	listeners map[uuid.UUID]func(value *T)
	kill      chan bool
}

// NewState creates a new instance of frontend state. It returns a pointer to the
// new state, with initial value, and a function to set the state.
func NewState[T any](component Component, value *T) (*State[T], func(*T)) {
	if &stateLock == nil {
		stateLock = sync.Mutex{}
	}
	logger.Log(DEBUG, "Creating new state, componentID: "+component.GetID().String())
	newState := &State[T]{
		Value:     value,
		id:        uuid.New(),
		ch:        make(chan *T),
		listeners: make(map[uuid.UUID]func(value *T)),
	}
	setState := func(newValue *T) {
		newState.ch <- newValue
	}
	registerKillChannel(component, newState)
	go listenForStateChange[T](component, newState)
	return newState, setState
}

// AddEffect adds an effect to the state. An effect is a function that is called
// whenever the state changes.
func (s *State[T]) AddEffect(effect func(value *T)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.listeners[uuid.New()] = effect
}

// listenForStateChange listens for state changes and updates the state accordingly.
func listenForStateChange[T any](component Component, state *State[T]) {
	logger.Log(DEBUG, "Listening for state change, componentID: "+component.GetID().String())
	for {
		select {
		case value := <-state.ch:
			state.lock.Lock()
			//println("State change detected, componentID: ", component.GetID().String())
			state.Value = value
			state.lock.Unlock()
			document.renderNotifier <- component
			notifyListeners[T](state)
		case <-state.kill:
			//println("Stopped listening for state change, componentID: ", component.GetID().String())
			// kill child components
			for _, child := range component.GetChildren() {
				killAllStates(child)
			}
			logger.Log(DEBUG, "Stopped listening for state change, componentID: "+component.GetID().String())
			return
		}
	}
}

// notifyListeners notifies all listeners of a state change.
func notifyListeners[T any](state *State[T]) {
	state.lock.Lock()
	defer state.lock.Unlock()
	for _, listener := range state.listeners {
		go listener(state.Value)
	}
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
			killAllStates(component)
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
				killAllStates((*input)[i])
			}
			*input = (*input)[:newLen]
		}
	}
}
