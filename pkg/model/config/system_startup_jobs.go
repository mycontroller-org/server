package config

// SystemStartupJobs config
type SystemStartupJobs struct {
	Restore StartupRestore `yaml:"restore"`
}

// StartupRestore loads data on startup
type StartupRestore struct {
	Enabled            bool   `yaml:"enabled"`
	ExtractedDirectory string `yaml:"extracted_directory"`
	ClearDatabase      bool   `yaml:"clean_database"`
}
