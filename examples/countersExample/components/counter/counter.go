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

	lowerID uuid.UUID
	raiseID uuid.UUID

	state    *goFE.State[counterState]
	setState func(*counterState)
	kill     chan bool
}

func NewCounter(props *Props) *Counter {
	newCounter := &Counter{
		id:      uuid.New(),
		lowerID: uuid.New(),
		raiseID: uuid.New(),
		kill:    make(chan bool),
	}
	newCounter.state, newCounter.setState = goFE.NewState[counterState](newCounter, &counterState{count: 0})
	return newCounter
}

func (c *Counter) GetID() uuid.UUID {
	return c.id
}

func (c *Counter) GetChildren() []goFE.Component {
	return nil
}

func (c *Counter) GetKill() chan bool {
	return c.kill
}

func (c *Counter) InitEventListeners() {
	goFE.GetDocument().AddEventListener(c.lowerID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("Clicked button")
		c.setState(&counterState{count: c.state.Value.count - 1})
		return nil
	}))
	goFE.GetDocument().AddEventListener(c.raiseID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("Clicked button")
		c.setState(&counterState{count: c.state.Value.count + 1})
		return nil
	}))
}

func (c *Counter) Render() string {
	return CounterTemplate(c.id.String(), c.state.Value.count, c.lowerID.String(), c.raiseID.String())
}
