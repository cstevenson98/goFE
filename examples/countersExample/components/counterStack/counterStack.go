//go:generate go run github.com/valyala/quicktemplate/qtc

package counterStack

import (
	"github.com/cstevenson98/goFE/examples/countersExample/components/counter"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	"math/rand"
	"syscall/js"
)

type Props struct {
	Title string
}

type counterStackState struct {
	numberOfCounters int
}

type CounterStack struct {
	id       uuid.UUID
	buttonID uuid.UUID
	props    Props
	state    *goFE.State[counterStackState]
	setState func(*counterStackState)
	counters []*counter.Counter
	kill     chan bool
}

const randCounterMax = 50

func NewApp(props Props) *CounterStack {
	randInt := rand.Intn(randCounterMax)
	var counters []*counter.Counter
	for i := 0; i < randInt; i++ {
		ctr := counter.NewCounter()
		counters = append(counters, ctr)
	}
	app := &CounterStack{
		id:       uuid.New(),
		buttonID: uuid.New(),
		props:    props,
		counters: counters,
		kill:     make(chan bool),
	}
	app.state, app.setState = goFE.NewState[counterStackState](app, &counterStackState{numberOfCounters: randInt})
	return app
}

func (a *CounterStack) GetID() uuid.UUID {
	return a.id
}

func (a *CounterStack) Render() string {
	goFE.UpdateComponentArray[*counter.Counter](&a.counters, a.state.Value.numberOfCounters, counter.NewCounter)
	var childrenResult []string
	for _, child := range a.counters {
		childrenResult = append(childrenResult, child.Render())
	}
	return CounterStackTemplate(a.id.String(), a.props.Title, childrenResult, a.buttonID.String())
}

func (a *CounterStack) GetChildren() []goFE.Component {
	var children []goFE.Component
	for _, child := range a.counters {
		children = append(children, child)
	}
	return children
}

func (a *CounterStack) GetKill() chan bool {
	return a.kill
}

func (a *CounterStack) InitEventListeners() {
	goFE.GetDocument().AddEventListener(a.buttonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.setState(&counterStackState{numberOfCounters: rand.Intn(randCounterMax)})
		return nil
	}))
}
