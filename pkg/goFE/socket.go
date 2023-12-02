package goFE

import (
	"errors"
	"syscall/js"
)

const closeNormalClosure = 1000

type Websocket struct {
	socket   js.Value
	url      string
	messages chan []byte
}

var ErrCouldNotGetWebsocket = errors.New("could not get websocket")

func NewWebsocket(url string) (*Websocket, error) {
	ws := js.Global().Get("WebSocket")
	if ws.IsNull() {
		return nil, ErrCouldNotGetWebsocket
	}

	return &Websocket{
		socket:   ws,
		url:      url,
		messages: make(chan []byte),
	}, nil
}

func (w *Websocket) Open() {
	w.socket = w.socket.New(w.url)
	println("Opened websocket")
	w.socket.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			println("Received message", args[0].Get("data").String())
			bytes := []byte(args[0].Get("data").String())
			w.messages <- bytes
		}()
		return this
	}))
}

func (w *Websocket) MessageChan() chan []byte {
	return w.messages
}

func (w *Websocket) Send(data string) {
	println("trying to send")
	w.socket.Call("send", data)
}

func (w *Websocket) Close() error {
	w.socket.Call("close", closeNormalClosure)
	return nil
}
