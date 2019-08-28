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
