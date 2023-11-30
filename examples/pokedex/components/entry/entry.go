//go:generate go run github.com/valyala/quicktemplate/qtc

package entry

import (
	"context"
	"encoding/json"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	fetch "marwan.io/wasm-fetch"
	"strconv"
	"time"
)

type Props struct {
	PokemonID int
}

type PokemonSprites struct {
	Other struct {
		OfficialArtwork struct {
			FrontDefault string `json:"front_default"`
		} `json:"official-artwork"`
	} `json:"other"`
}

type Pokemon struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	Sprites PokemonSprites `json:"sprites"`
}

type Entry struct {
	id uuid.UUID

	pokemon    *goFE.State[Pokemon]
	setPokemon func(*Pokemon)
}

func NewEntry(props *Props) *Entry {
	entry := &Entry{
		id: uuid.New(),
	}
	entry.pokemon, entry.setPokemon = goFE.NewState[Pokemon](entry, nil)
	go func() {
		if props != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			res, err := fetch.Fetch("https://pokeapi.co/api/v2/pokemon/"+strconv.Itoa(props.PokemonID), &fetch.Opts{
				Method: fetch.MethodGet,
				Signal: ctx,
			})
			if err != nil {
				println(err.Error())
			}
			var pokemon Pokemon
			err = json.Unmarshal(res.Body, &pokemon)
			if err != nil {
				println(err.Error())
			}
			entry.setPokemon(&pokemon)
		}
	}()
	return entry
}

func (e *Entry) GetID() uuid.UUID {
	return e.id
}

func (e *Entry) Render() string {
	return EntryTemplate(e.id.String(), e.pokemon.Value)
}

func (e *Entry) GetChildren() []goFE.Component {
	return nil
}

func (e *Entry) InitEventListeners() {}
