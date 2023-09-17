//go:generate go run github.com/valyala/quicktemplate/qtc

package entry

import (
	"context"
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	fetch "marwan.io/wasm-fetch"
	"time"
)

type Props struct {
}

type entryState struct {
}

type Entry struct {
	id   uuid.UUID
	kill chan bool
}

func NewEntry(props Props) *Entry {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := fetch.Fetch("https://pokeapi.co/api/v2/pokemon/?limit=25&offset=0", &fetch.Opts{
			Method: fetch.MethodGet,
			Signal: ctx,
		})
		if err != nil {
			println(err.Error())
		}
		print(string(res.Body))
	}()

	return &Entry{
		id:   uuid.New(),
		kill: make(chan bool),
	}
}

func (e *Entry) GetID() uuid.UUID {
	return e.id
}

func (e *Entry) Render() string {
	return EntryTemplate(e.id.String())
}

func (e *Entry) GetChildren() []goFE.Component {
	return nil
}

func (e *Entry) GetKill() chan bool {
	return e.kill
}

func (e *Entry) InitEventListeners() {}
