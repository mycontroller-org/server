package config

import "github.com/mycontroller-org/backend/v2/pkg/model/cmap"

// Config of the system
type Config struct {
	Web         WebConfig                `yaml:"web"`
	Directories Directories              `yaml:"directories"`
	Logger      LoggerConfig             `yaml:"logger"`
	Secret      string                   `yaml:"secret"` // secret used to encrypt sensitive data
	Bus         cmap.CustomMap           `yaml:"bus"`
	Gateway     cmap.CustomMap           `yaml:"gateway"`
	Database    Database                 `yaml:"database"`
	Databases   []map[string]interface{} `yaml:"databases"`
	StartupJobs Startup                  `yaml:"startup_jobs"`
}

// WebConfig input
type WebConfig struct {
	BindAddress     string `yaml:"bindAddress"`
	Port            uint   `yaml:"port"`
	WebDirectory    string `yaml:"webDirectory"`
	EnableProfiling bool   `yaml:"enable_profiling"`
}

// Directories for data and logs
type Directories struct {
	Data string `yaml:"data"`
	Logs string `yaml:"logs"`
}

// LoggerConfig input
type LoggerConfig struct {
	Mode     string         `yaml:"mode"`
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

// Startup jobs
type Startup struct {
	Importer StartupImporter `yaml:"importer"`
}

// StartupImporter loads data on startup
type StartupImporter struct {
	Enabled         bool   `yaml:"enabled"`
	TargetDirectory string `yaml:"target_directory"`
	Type            string `yaml:"type"`
	ClearDatabase   bool   `yaml:"clean_database"`
}
