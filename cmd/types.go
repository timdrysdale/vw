package cmd

import (
	"sync"

	"github.com/gobwas/ws"
)

type Commands struct {
	Commands []string `yaml:"commands"`
}

type Endpoints map[string]string

type Stream struct {
	Destination string      `yaml:"destination"`
	Feeds       interface{} `yaml:"feeds"`
	InputNames  []string
	//InputChannels []chan Packet
}

type Output struct {
	Streams []Stream `yaml:"streams"`
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

//config:
//  control:
//    path: control
//    scheme: http
//  host: "wss://video.practable.io:443"
//  log: ./vw.log
//  retry_wait: 1000
//  strict: false
//  tuning:
//    bufferSize: 1024000
//  uuid: 49270598-9da2-4209-98da-e559f0c587b4
//  verbose: false
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
