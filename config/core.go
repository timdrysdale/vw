package config

type CoreConfiguration struct {
	Mux        MuxConfiguration `mapstructure:"mux"`
	Logfile    string           `mapstructure:"logfile"`
	Debug      bool             `mapstructure:"debug"`
	Verbose    bool             `mapstructure:"verbose"`
	CPUProfile string           `mapstructure:"cpuprofile"`
	MemProfile string           `mapstructure:"memprofile"`
}

type MuxConfiguration struct {
	WorkerCount int `mapstructure:"workerCount"`
	BufferSize  int `mapstructure:"bufferSize"`
}
