package cmd

import (
	"bytes"
	"sync"

	"github.com/gobwas/ws"
)

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

type FeedMap map[string][]chan Packet

type ClientMap map[string][]chan Packet

type ChannelDetails struct {
	Channel     chan Packet
	Feed        string
	Destination string
}

type Variables struct {
	Vars map[string]string `yaml:"variables"`
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

// Types from crossbar

// messages will be wrapped in this struct for muxing
type message struct {
	sender clientDetails
	op     ws.OpCode
	data   []byte //text data are converted to/from bytes as needed
}

type clientDetails struct {
	name         string
	topic        string
	messagesChan chan message
}

// requests to add or delete subscribers are represented by this struct
type clientAction struct {
	action clientActionType
	client clientDetails
}

// userActionType represents the type of of action requested
type clientActionType int

// clientActionType constants
const (
	clientAdd clientActionType = iota
	clientDelete
)

type topicDirectory struct {
	sync.Mutex
	directory map[string][]clientDetails
}

// gobwas/ws
type readClientDataReturns struct {
	msg []byte
	op  ws.OpCode
	err error
}

type mutexBuffer struct {
	mux sync.Mutex
	b   bytes.Buffer
}
