package cmd

import (
	"bytes"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/hub"
	"github.com/timdrysdale/rwc"
)

type App struct {
	Closed    chan struct{}
	Hub       *agg.Hub
	Opts      Specification
	Websocket *rwc.Hub
	WaitGroup sync.WaitGroup
}

type WsHandlerClient struct {
	Messages   *hub.Client
	Conn       *websocket.Conn
	UserAgent  string //r.UserAgent()
	RemoteAddr string //r.Header.Get("X-Forwarded-For")
}

type mutexBuffer struct {
	mux sync.Mutex
	b   bytes.Buffer
}

type Packet struct {
	Data []byte
}
