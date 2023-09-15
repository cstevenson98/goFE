//go:generate go run github.com/valyala/quicktemplate/qtc

package app

import (
	"github.com/cstevenson98/goFE/internal/components/counter"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	"math/rand"
	"syscall/js"
)

type Props struct {
	Title string
}

type appState struct {
	numberOfCounters int
}

type App struct {
	id       uuid.UUID
	buttonID uuid.UUID
	props    Props
	state    *goFE.State[appState]
	setState func(*appState)
	counters []*counter.Counter
	kill     chan bool
}

const randCounterMax = 10

func NewApp(props Props) *App {
	randInt := randCounterMax
	noOfCounter, setNoOfCounter := goFE.NewState[appState](&appState{numberOfCounters: randInt})

	// Make a bunch of counters
	var children []*counter.Counter
	for i := 0; i < randInt; i++ {
		ctr := counter.NewCounter()
		children = append(children, ctr)
	}

	app := &App{
		id:       uuid.New(),
		buttonID: uuid.New(),
		props:    props,
		state:    noOfCounter,
		setState: setNoOfCounter,
		counters: children,
	}

	go goFE.ListenForStateChange[appState](app, noOfCounter)
	return app
}

func (a *App) GetID() uuid.UUID {
	return a.id
}

func (a *App) Render() string {
	goFE.UpdateStateArray[*counter.Counter](&a.counters, a.state.Value.numberOfCounters, counter.NewCounter)
	var childrenResult []string
	for _, child := range a.counters {
		childrenResult = append(childrenResult, child.Render())
	}
	return AppTemplate(a.id.String(), a.props.Title, childrenResult, a.buttonID.String())
}

func (a *App) GetChildren() []goFE.Component {
	var children []goFE.Component
	for _, child := range a.counters {
		children = append(children, child)
	}
	return children
}

func (a *App) GetKill() chan bool {
	return a.kill
}

func (a *App) InitEventListeners() {
	goFE.GetDocument().AddEventListener(a.buttonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.setState(&appState{numberOfCounters: rand.Intn(randCounterMax)})
		return nil
	}))
}
