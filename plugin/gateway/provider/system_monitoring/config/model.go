package config

// Config data
type Config struct {
	HostIDMap     map[string]string
	HostConfigMap map[string]HostConfig
}

type HostConfig struct {
	Disabled    bool
	Memory      MemoryConfig
	CPU         CPUConfig
	Disk        DiskConfig
	Temperature TemperatureConfig
	Process     ProcessConfig
}

// Memory details
type MemoryConfig struct {
	Interval           string
	MemoryDisabled     bool
	SwapMemoryDisabled bool
}

// CPU details
type CPUConfig struct {
	Interval       string
	CPUDisabled    bool
	PerCPUDisabled bool
}

// Disk details
type DiskConfig struct {
	Interval string
	Disabled bool
	Data     map[string]DiskData
}

// DiskData struct
type DiskData struct {
	Disabled bool
	Name     string
	Path     string
}

// Temperature details
type TemperatureConfig struct {
	Interval    string
	DisabledAll bool
	Enabled     []string
}

// Process details
type ProcessConfig struct {
	Interval string
	Disabled bool
	Data     map[string]ProcessData
}

// ProcessData struct
type ProcessData struct {
	Disabled bool
	Name     string
	Filter   map[string]string
}

// Keys
const (
	SourceTypeMemory      = "memory"
	SourceTypeSwapMemory  = "swap_memory"
	SourceTypeCPU         = "cpu"
	SourceTypeTemperature = "temperature"
	SourceTypeDisk        = "disk"
	SourceTypeProcess     = "process"
)
