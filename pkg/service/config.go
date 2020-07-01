package service

// Database to be used
type Database struct {
	Storage string `yaml:"storage"`
	Metrics string `yaml:"metrics"`
}

// WebConfig input
type WebConfig struct {
	BindAddress  string `yaml:"bindAddress"`
	Port         uint   `yaml:"port"`
	WebDirectory string `yaml:"webDirectory"`
}

// Config of the system
type Config struct {
	Web       WebConfig           `yaml:"web"`
	Database  Database            `yaml:"database"`
	Databases []map[string]string `yaml:"databases"`
	Logger    map[string]string   `yaml:"logger"`
}
