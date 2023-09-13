//go:generate go run github.com/valyala/quicktemplate/qtc

package counter

import (
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
)

type Props struct{}

type counterState struct {
	count int
}

type Counter struct {
	id    uuid.UUID
	props Props

	state goFE.State[counterState]
	kill  chan bool
}

func NewCounter() *Counter {
	count := goFE.State[counterState]{
		Value: counterState{
			count: 0,
		},
		Chan: make(chan *counterState),
	}

	return &Counter{
		id:    uuid.New(),
		state: count,
		kill:  make(chan bool),
	}
}

func (b *Counter) GetID() uuid.UUID {
	return b.id
}

func (b *Counter) Render() string {
	return CounterTemplate(b.id.String(), b.state.Value.count)
}

func (b *Counter) GetChildren() []goFE.Component {
	return nil
}

func (b *Counter) GetKill() chan bool {
	return b.kill
}
