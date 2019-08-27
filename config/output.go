package config

/*
// StreamsConfiguration holds an array of Stream structs
//
type OutputConfiguration struct {
	StreamList Streams `yaml:"streams"`
}

//type CaptureConfiguration struct {
//	CommandList Commands `yaml:"commands"`
//}
//
//type Commands struct {
//	Commands []string
//}

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
	From       interface{} `mapstructure:"from"`
	InputNames []string
	Serve      string `mapstructure:"serve"`
	To         string `mapstructure:"to"`
}

type Streams struct {
	Streams []StreamDetails
}
*/
