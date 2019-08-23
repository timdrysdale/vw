package cmd

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

type ClientDetails struct {
	Destination string
	Channels    []chan Packet
}

type ClientList []ClientDetails

type Everything struct {
	ClientList ClientList
	FeedMap    FeedMap
	Endpoints  Endpoints
	Output     Output
}

type ChannelDetails struct {
	Channel     chan Packet
	Feed        string
	Destination string
}
