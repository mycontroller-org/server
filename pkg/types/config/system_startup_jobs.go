package config

// SystemStartupJobs config
type SystemStartupJobs struct {
	Restore StartupRestore `yaml:"restore"`
}

// StartupRestore loads data on startup
type StartupRestore struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	ExtractedDirectory string `json:"extracted_directory" yaml:"extracted_directory"`
	ClearDatabase      bool   `json:"clean_database" yaml:"clean_database"`
}
