package config

// Config of the system
type Config struct {
	Web       WebConfig                `yaml:"web"`
	Logger    LoggerConfig             `yaml:"logger"`
	Database  Database                 `yaml:"database"`
	Databases []map[string]interface{} `yaml:"databases"`
}

// WebConfig input
type WebConfig struct {
	BindAddress  string `yaml:"bindAddress"`
	Port         uint   `yaml:"port"`
	WebDirectory string `yaml:"webDirectory"`
}

// LoggerConfig input
type LoggerConfig struct {
	Encoding string         `yaml:"encoding"`
	Level    LogLevelConfig `yaml:"level"`
}

// LogLevelConfig input
type LogLevelConfig struct {
	Core    string `yaml:"core"`
	Storage string `yaml:"storage"`
	Metrics string `yaml:"metrics"`
}

// Database to be used
type Database struct {
	Storage string `yaml:"storage"`
	Metrics string `yaml:"metrics"`
}
