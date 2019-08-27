//Config has structs to allow the YAML configuration file to be
//unmarshalled by viper
package config

//Configuration is the top-level struct
//
//Each sub-level has its own struct
type Configuration struct {
	Control ControlConfiguration `yaml:"control"`
	Capture CaptureConfiguration `yaml:"capture"`
	Core    CoreConfiguration    `yaml:"core"`
	Collect CollectConfiguration `yaml:"collect"`
	Output  OutputConfiguration  `yaml:"output"`
	Extra   ExtraConfiguration   `yaml:"extra"`
}
