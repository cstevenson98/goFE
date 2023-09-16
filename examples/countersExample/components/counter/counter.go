//go:generate go run github.com/valyala/quicktemplate/qtc

package counter

import (
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	"syscall/js"
)

type Props struct{}

type counterState struct {
	count int
}

type Counter struct {
	id    uuid.UUID
	props Props

	state    *goFE.State[counterState]
	setState func(*counterState)
	kill     chan bool
}

func NewCounter() *Counter {
	newCounter := &Counter{
		id:   uuid.New(),
		kill: make(chan bool),
	}
	newCounter.state, newCounter.setState = goFE.NewState[counterState](newCounter, &counterState{count: 0})
	return newCounter
}

func (b *Counter) GetID() uuid.UUID {
	return b.id
}

func (b *Counter) GetChildren() []goFE.Component {
	return nil
}

func (b *Counter) GetKill() chan bool {
	return b.kill
}

func (b *Counter) InitEventListeners() {
	goFE.GetDocument().AddEventListener(b.id, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("Clicked button")
		b.setState(&counterState{count: b.state.Value.count + 1})
		return nil
	}))
}

func (b *Counter) Render() string {
	return CounterTemplate(b.id.String(), b.state.Value.count)
}
