package cmd

import (
	"bytes"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/timdrysdale/hub"
)

//definitely used

type WsHandlerClient struct {
	Messages   *hub.Client
	Conn       *websocket.Conn
	UserAgent  string //r.UserAgent()
	RemoteAddr string //r.Header.Get("X-Forwarded-For")
}

type ClientOptions struct {
	BufferLength int `yaml:"bufferLength"`
	TimeoutMS    int `yaml:"timeoutMS"`
}

type MuxOptions struct {
	BufferSize int `yaml:"bufferSize"`
	Workers    int `yaml:"workers"`
}

type HTTPOptions struct {
	Port      int `yaml:"port"`
	WaitMS    int `yaml:"waitMS"`
	FlushMS   int `yaml:"flushMS"`
	TimeoutMS int `yaml:"timeoutMS"`
}

// Might be used, might not be

type Commands struct {
	Commands []string `yaml:"commands"`
}

type Endpoints map[string]string

type Monitor struct {
	Monitor []string `yaml:"monitor"`
}

type Stream struct {
	Destination string      `yaml:"destination"`
	Feeds       interface{} `yaml:"feeds"`
	Local       string      `yaml:"local"`
	InputNames  []string
}

type Output struct {
	Streams []Stream `yaml:"streams"`
}

type Writer struct {
	File       string      `yaml:"file"`
	Feeds      interface{} `yaml:"feeds"`
	InputNames []string
	Debug      bool `yaml:"debug"`
}

type ToFile struct {
	Writers []Writer `yaml:"writers"`
}

type Packet struct {
	Data []byte
}

//??
type mutexBuffer struct {
	mux sync.Mutex
	b   bytes.Buffer
}
