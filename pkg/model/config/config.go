package config

import "github.com/mycontroller-org/backend/v2/pkg/model/cmap"

const (
	SystemStartJobsFilename = "system_startup_jobs.yaml"
	UserStartJobsFilename   = "user_startup_jobs.yaml"
)

// Config of the system
type Config struct {
	Web         WebConfig                `yaml:"web"`
	Directories Directories              `yaml:"directories"`
	Logger      LoggerConfig             `yaml:"logger"`
	Secret      string                   `yaml:"secret"` // secret used to encrypt sensitive data
	Bus         cmap.CustomMap           `yaml:"bus"`
	Gateway     cmap.CustomMap           `yaml:"gateway"`
	Handler     cmap.CustomMap           `yaml:"handler"`
	Task        cmap.CustomMap           `yaml:"task"`
	Database    Database                 `yaml:"database"`
	Databases   []map[string]interface{} `yaml:"databases"`
}

// WebConfig input
type WebConfig struct {
	WebDirectory     string          `yaml:"web_directory"`
	DocumentationURL string          `yaml:"documentation_url"`
	EnableProfiling  bool            `yaml:"enable_profiling"`
	Http             HttpConfig      `yaml:"http"`
	HttpsSSL         HttpsSSLConfig  `yaml:"https_ssl"`
	HttpsACME        HttpsACMEConfig `yaml:"https_acme"`
}

// HttpConfig struct
type HttpConfig struct {
	Enabled     bool   `yaml:"enabled"`
	BindAddress string `yaml:"bind_address"`
	Port        uint   `yaml:"port"`
}

// HttpsSSLConfig struct
type HttpsSSLConfig struct {
	Enabled     bool   `yaml:"enabled"`
	BindAddress string `yaml:"bind_address"`
	Port        uint   `yaml:"port"`
	CertDir     string `yaml:"cert_dir"`
}

// HttpsACMEConfig struct
type HttpsACMEConfig struct {
	Enabled       bool     `yaml:"enabled"`
	BindAddress   string   `yaml:"bind_address"`
	Port          uint     `yaml:"port"`
	CacheDir      string   `yaml:"cache_dir"`
	ACMEDirectory string   `yaml:"acme_directory"`
	Email         string   `yaml:"email"`
	Domains       []string `yaml:"domains"`
}

// Directories for data and logs
type Directories struct {
	Data          string `yaml:"data"`
	Logs          string `yaml:"logs"`
	Tmp           string `yaml:"tmp"`
	SecureShare   string `yaml:"secure_share"`
	InsecureShare string `yaml:"insecure_share"`
}

// LoggerConfig input
type LoggerConfig struct {
	Mode     string         `yaml:"mode"`
	Encoding string         `yaml:"encoding"`
	Level    LogLevelConfig `yaml:"level"`
}

// LogLevelConfig input
type LogLevelConfig struct {
	Core       string `yaml:"core"`
	WebHandler string `yaml:"web_handler"`
	Storage    string `yaml:"storage"`
	Metrics    string `yaml:"metrics"`
}

// Database to be used
type Database struct {
	Storage string `yaml:"storage"`
	Metrics string `yaml:"metrics"`
}
