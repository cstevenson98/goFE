//go:generate go run github.com/valyala/quicktemplate/qtc

package app

import (
	"github.com/cstevenson98/goFE/internal/components/counter"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
)

type Props struct {
	Title string
}

type App struct {
	id       uuid.UUID
	props    Props
	children []goFE.Component

	kill chan bool
}

func NewApp(props Props) *App {
	ctr := counter.NewCounter()
	children := []goFE.Component{ctr}
	return &App{
		id:       uuid.New(),
		props:    props,
		children: children,
	}
}

func (a *App) GetID() uuid.UUID {
	return a.id
}

func (a *App) Render() string {
	var childrenResult []string
	for _, child := range a.children {
		childrenResult = append(childrenResult, child.Render())
	}
	return AppTemplate(a.props.Title, childrenResult)
}

func (a *App) GetChildren() []goFE.Component {
	return a.children
}

func (a *App) GetKill() chan bool {
	return a.kill
}

func (a *App) InitEventListeners() {}
