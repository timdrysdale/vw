package config

type CoreConfiguration struct {
	Mux        MuxConfiguration `yaml:"mux"`
	Logfile    string           `yaml:"logfile"`
	Debug      bool             `yaml:"debug"`
	Verbose    bool             `yaml:"verbose"`
	CPUProfile string           `yaml:"cpuprofile"`
	MemProfile string           `yaml:"memprofile"`
}

type MuxConfiguration struct {
	WorkerCount int `yaml:"workerCount"`
	BufferSize  int `yaml:"bufferSize"`
}
