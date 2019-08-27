package config

import (
	"sync"

	"github.com/gobwas/ws"
)

type Commands struct {
	Commands []string `yaml:"commands"`
}

//Stream struct holds details of outgoing streams
//Note that InputNames holds reformatted versions of the
//feed names read out of the "from" field in the config
//
//Example config YAML:
// ---
// streams:
//   -   to: "${relay}/${uuid}/${session}/front/medium"
//   	  serve: "${localhost}/front/medium"
//       from:
//         - audio
//         - videoFrontMedium
type StreamDetails struct {
	From       interface{} `yaml:"from"`
	InputNames []string
	Serve      string `yaml:"serve"`
	To         string `yaml:"to"`
}

type Streams struct {
	Stream []StreamDetails `yaml:"streams"`
}

type Endpoints map[string]string

type Packet struct {
	Data []byte
}

type FeedMap map[string][]chan Packet

type ClientMap map[string][]chan Packet

type ChannelDetails struct {
	Channel chan Packet
	From    string
	To      string
}

type Message struct {
	Sender ClientDetails
	Op     ws.OpCode
	Data   []byte //text data are converted to/from bytes as needed
}

type ClientDetails struct {
	Name         string
	Topic        string
	MessagesChan chan Message
}

// requests to add or delete subscribers are represented by this struct
type ClientAction struct {
	Action ClientActionType
	Client ClientDetails
}

// userActionType represents the type of of action requested
type ClientActionType int

// clientActionType constants
const (
	ClientAdd ClientActionType = iota
	ClientDelete
)

type TopicDirectory struct {
	sync.Mutex
	Directory map[string][]ClientDetails
}

// gobwas/ws
type ReadClientDataReturns struct {
	Msg []byte
	Op  ws.OpCode
	Err error
}
