//go:generate go run github.com/valyala/quicktemplate/qtc
package pokedex

import (
	"github.com/cstevenson98/goFE/examples/pokedex/components/entry"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
)

type Props struct{}

type pokedexState struct {
	listResult PokemonResultsList
}

type Pokedex struct {
	id       uuid.UUID
	kill     chan bool
	state    *goFE.State[pokedexState]
	setState func(*pokedexState)

	entries []*entry.Entry
}

const (
	initialOffset = 0
	initialLimit  = 25
)

func NewPokedex(props Props) *Pokedex {
	pokedex := &Pokedex{
		id:   uuid.New(),
		kill: make(chan bool),
	}

	for i := initialOffset; i < initialOffset+initialLimit; i++ {
		println(i)
		pokedex.entries = append(pokedex.entries, entry.NewEntry(&entry.Props{
			PokemonID: i + 1,
		}))
	}

	pokedex.state, pokedex.setState = goFE.NewState[pokedexState](pokedex, &pokedexState{})
	return pokedex
}

func (p *Pokedex) GetID() uuid.UUID {
	return p.id
}

func (p *Pokedex) Render() string {
	goFE.UpdateComponentArray[*entry.Entry, entry.Props](&p.entries, initialLimit, entry.NewEntry, nil)
	return PokedexTemplate(p.id.String(), goFE.RenderChildren(p))
}

func (p *Pokedex) GetChildren() []goFE.Component {
	var out []goFE.Component
	for _, child := range p.entries {
		out = append(out, child)
	}
	return out
}

func (p *Pokedex) GetKill() chan bool {
	return p.kill
}

func (p *Pokedex) InitEventListeners() {}
