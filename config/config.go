//Config has structs to allow the YAML configuration file to be
//unmarshalled by viper
package config

//Configuration is the top-level struct
//
//Each sub-level has its own struct

type Capture struct {
	Commands []string
}

type Commands struct {
	Commands []string
}
type Control struct {
	Parameters map[string]string
}
type Collect struct {
	Parameters map[string]string
}
type Extra struct {
	Parameters map[string]string
}

//type Core struct {
//	parameters map[string]string
//}

type Output struct {
	Streams []StreamDetails
}

type Configuration struct {
	Control `mapstructure:",squash"`
	Capture `mapstructure:",squash"`
	Core    map[string]string
	Collect `mapstructure:",squash"`
	Output  `mapstructure:",squash"`
	Extra   `mapstructure:",squash"`
}

//type Configuration struct {
//	Control ControlConfiguration `mapstructure:",squash"`
//	Capture CaptureConfiguration `mapstructure:",squash"`
//	Core    CoreConfiguration    `mapstructure:",squash"`
//	Collect CollectConfiguration `mapstructure:",squash"`
//	Output  OutputConfiguration  `mapstructure:",squash"`
//	Extra   ExtraConfiguration   `mapstructure:",squash"`
//}
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
