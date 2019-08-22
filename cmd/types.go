package cmd

//type Stream struct {
//	Destination string
//	Feeds       []string
//}

type Stream struct {
	Name          string
	Destination   string
	InputNames    []string `yaml:"feeds"`
	InputChannels []chan Packet
}

type Output struct {
	Streams []Stream `yaml:"streams"`
}

type Packet struct {
	Data []byte
}
