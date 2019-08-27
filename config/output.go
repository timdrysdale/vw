package config

// StreamsConfiguration holds an array of Stream structs
//
type OutputConfiguration struct {
	Streams Streams `yaml:"streams"`
}
