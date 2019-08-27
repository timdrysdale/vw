package cmd

type Commands struct {
	Commands []string `yaml:"commands"`
}

type Endpoints map[string]string

type Stream struct {
	Destination string      `yaml:"destination"`
	Feeds       interface{} `yaml:"feeds"`
	Local       string      `yaml:"local"`
	InputNames  []string
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

type Variables struct {
	Vars map[string]string `yaml:"variables"`
}

type ClientOptions struct {
	BufferSize int `yaml:"bufferSize"`
}

type MuxOptions struct {
	BufferSize int `yaml:"bufferSize"`
	Workers    int `yaml:"workers"`
}

type HTTPOptions struct {
	Port   int `yaml:"port"`
	WaitMS int `yaml:"waitMS"`
}
