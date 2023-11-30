//go:generate go run github.com/valyala/quicktemplate/qtc
package pokedex

import (
	"context"
	"encoding/json"
	"github.com/cstevenson98/goFE/examples/pokedex/components/entry"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	fetch "marwan.io/wasm-fetch"
	"strings"
	"syscall/js"
	"time"
)

type Props struct{}

type pokedexState struct {
	allPokemon PokemonResultsList
}

func FilterResultByName(name *string, list PokemonResultsList) []int {
	var out []int
	for i, result := range list.Results {
		if name == nil || strings.Contains(result.Name, *name) {
			out = append(out, i+1)
		}
	}
	return out
}

type Pokedex struct {
	id               uuid.UUID
	formID           uuid.UUID
	inputID          uuid.UUID
	state            *goFE.State[pokedexState]
	setState         func(*pokedexState)
	searchTerm       *goFE.State[string]
	setSearchTerm    func(*string)
	searchResults    *goFE.State[[]int]
	setSearchResults func(*[]int)

	inputValue string

	entries []*entry.Entry
}

const (
	initialOffset = 0
	initialLimit  = 25
)

func NewPokedex(_ Props) *Pokedex {
	pokedex := &Pokedex{
		id:      uuid.New(),
		formID:  uuid.New(),
		inputID: uuid.New(),
	}

	for i := initialOffset; i < initialOffset+initialLimit; i++ {
		pokedex.entries = append(pokedex.entries, entry.NewEntry(&entry.Props{
			PokemonID: i + 1,
		}))
	}

	pokedex.state, pokedex.setState = goFE.NewState[pokedexState](pokedex, &pokedexState{})
	pokedex.searchResults, pokedex.setSearchResults = goFE.NewState[[]int](pokedex, &[]int{})
	pokedex.searchTerm, pokedex.setSearchTerm = goFE.NewState[string](pokedex, nil)
	pokedex.searchTerm.AddEffect(func(value *string) {
		indices := FilterResultByName(value, pokedex.state.Value.allPokemon)
		pokedex.setSearchResults(&indices)
	})
	pokedex.state.AddEffect(func(value *pokedexState) {
		indices := FilterResultByName(pokedex.searchTerm.Value, value.allPokemon)
		pokedex.setSearchResults(&indices)
	})
	go func() { // Async fetch of all pokemon
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res, err := fetch.Fetch(ListPokemonURL(-1, 0), &fetch.Opts{
			Method: fetch.MethodGet,
			Signal: ctx,
		})
		if err != nil {
			println(err.Error())
		}
		var listResult PokemonResultsList
		err = json.Unmarshal(res.Body, &listResult)
		if err != nil {
			println(err.Error())
		}
		pokedex.setState(&pokedexState{allPokemon: listResult})
	}()
	return pokedex
}

func (p *Pokedex) GetID() uuid.UUID {
	return p.id
}

func (p *Pokedex) Render() string {
	var newProps []*entry.Props
	if p.searchResults.Value != nil {
		for i, index := range *p.searchResults.Value {
			newProps = append(newProps, &entry.Props{
				PokemonID: index,
			})
			if i >= initialLimit {
				break
			}
		}
	}
	goFE.UpdateComponentArray[*entry.Entry, entry.Props](&p.entries, len(newProps), entry.NewEntry, newProps)
	value := p.searchTerm.Value
	if value == nil {
		newValue := ""
		value = &newValue
	}
	return PokedexTemplate(p.id.String(), p.formID.String(), p.inputID.String(), *value, goFE.RenderChildren(p))
}

func (p *Pokedex) GetChildren() []goFE.Component {
	var out []goFE.Component
	for _, child := range p.entries {
		out = append(out, child)
	}
	return out
}

func (p *Pokedex) InitEventListeners() {
	goFE.GetDocument().AddEventListener(p.formID, "submit", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		p.setSearchTerm(&p.inputValue)
		return nil
	}))
	goFE.GetDocument().AddEventListener(p.inputID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		p.inputValue = this.Get("value").String()
		return nil
	}))
}
