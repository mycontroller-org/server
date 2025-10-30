package config

import (
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
)

const (
	SystemStartJobsFilename = "system_startup_jobs.yaml"
	UserStartJobsFilename   = "user_startup_jobs.yaml"
)

// Config of the system
type Config struct {
	Secret           string             `yaml:"secret"`   // secret used to encrypt sensitive data
	JwtSeed          string             `yaml:"jwt_seed"` // optional static seed used when deriving JWT secret, otherwise host-id is used
	Telemetry        TelemetryConfig    `yaml:"telemetry"`
	Web              WebConfig          `yaml:"web"`
	Logger           LoggerConfig       `yaml:"logger"`
	Directories      Directories        `yaml:"directories"`
	Bus              cmap.CustomMap     `yaml:"bus"`
	Gateway          sfTY.ServiceFilter `yaml:"gateway"`
	Handler          sfTY.ServiceFilter `yaml:"handler"`
	Task             sfTY.ServiceFilter `yaml:"task"`
	Schedule         sfTY.ServiceFilter `yaml:"schedule"`
	VirtualAssistant sfTY.ServiceFilter `yaml:"virtual_assistant"`
	Database         Database           `yaml:"database"`
}

// WebConfig input
type WebConfig struct {
	WebDirectory     string          `yaml:"web_directory"`
	DocumentationURL string          `yaml:"documentation_url"`
	EnableProfiling  bool            `yaml:"enable_profiling"`
	ReadTimeout      string          `yaml:"read_timeout"`
	Http             HttpConfig      `yaml:"http"`
	HttpsSSL         HttpsSSLConfig  `yaml:"https_ssl"`
	HttpsACME        HttpsACMEConfig `yaml:"https_acme"`
}

// TelemetryConfig input
type TelemetryConfig struct {
	Enabled bool `yaml:"enabled"`
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
	Mode             string         `yaml:"mode"`
	Encoding         string         `yaml:"encoding"`
	Level            LogLevelConfig `yaml:"level"`
	EnableStacktrace bool           `yaml:"enable_stacktrace"`
}

// LogLevelConfig input
type LogLevelConfig struct {
	Core       string `yaml:"core"`
	WebHandler string `yaml:"web_handler"`
	Storage    string `yaml:"storage"`
	Metric     string `yaml:"metric"`
}

// Database to be used
type Database struct {
	Storage cmap.CustomMap `yaml:"storage"`
	Metric  cmap.CustomMap `yaml:"metric"`
}

func (c *Config) Clone() *Config {
	cloned := Config{
		Secret:           c.Secret,
		Telemetry:        c.Telemetry,
		Web:              c.Web,
		Logger:           c.Logger,
		Directories:      c.Directories,
		Bus:              c.Bus.Clone(),
		Gateway:          *c.Gateway.Clone(),
		Handler:          *c.Handler.Clone(),
		Task:             *c.Task.Clone(),
		Schedule:         *c.Schedule.Clone(),
		VirtualAssistant: *c.VirtualAssistant.Clone(),
		Database:         c.Database,
	}

	return &cloned
}
