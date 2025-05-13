//go:generate go run github.com/valyala/quicktemplate/qtc
package messageBoard

import (
	"bytes"
	"context"
	"encoding/json"
	"syscall/js"
	"time"

	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	fetch "marwan.io/wasm-fetch"
)

type Props struct{}

type Message struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type messageBoardState struct {
	Messages []Message
}

type MessageBoard struct {
	id       uuid.UUID
	formID   uuid.UUID
	inputID  uuid.UUID
	state    *goFE.State[messageBoardState]
	setState func(*messageBoardState)
}

func NewMessageBoard(_ Props) *MessageBoard {
	mb := &MessageBoard{
		id:      uuid.New(),
		formID:  uuid.New(),
		inputID: uuid.New(),
	}
	mb.state, mb.setState = goFE.NewState[messageBoardState](mb, &messageBoardState{})

	// Initial fetch of messages
	mb.fetchMessages()

	return mb
}

func (mb *MessageBoard) fetchMessages() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := fetch.Fetch("/api/messages", &fetch.Opts{
			Method: fetch.MethodGet,
			Signal: ctx,
		})
		if err != nil {
			println("Error fetching messages:", err.Error())
			return
		}

		var messages []Message
		err = json.Unmarshal(res.Body, &messages)
		if err != nil {
			println("Error parsing messages:", err.Error())
			return
		}

		mb.setState(&messageBoardState{Messages: messages})
	}()
}

func (mb *MessageBoard) GetID() uuid.UUID {
	return mb.id
}

func (mb *MessageBoard) GetChildren() []goFE.Component {
	return nil
}

func (mb *MessageBoard) InitEventListeners() {
	// Handle form submission
	goFE.GetDocument().AddEventListener(mb.formID, "submit", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")

		// Get input value
		input := js.Global().Get("document").Call("getElementById", mb.inputID.String())
		content := input.Get("value").String()

		if content == "" {
			return nil
		}

		// Post new message
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			body, err := json.Marshal(map[string]string{"content": content})
			if err != nil {
				println("Error marshaling request:", err.Error())
				return
			}

			_, err = fetch.Fetch("/api/messages", &fetch.Opts{
				Method:  fetch.MethodPost,
				Body:    bytes.NewReader(body),
				Headers: map[string]string{"Content-Type": "application/json"},
				Signal:  ctx,
			})
			if err != nil {
				println("Error posting message:", err.Error())
				return
			}

			// Clear input
			input.Set("value", "")

			// Refresh messages
			mb.fetchMessages()
		}()

		return nil
	}))
}

func (mb *MessageBoard) Render() string {
	messages := mb.state.Value.Messages
	return MessageBoardTemplate(mb.id.String(), mb.formID.String(), mb.inputID.String(), messages)
}
