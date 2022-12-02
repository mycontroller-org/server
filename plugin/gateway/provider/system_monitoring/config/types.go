package config

// Config data
type Config struct {
	HostConfigMap map[string]HostConfig `json:"hostConfigMap"`
}

type HostConfig struct {
	Disabled    bool              `json:"disabled" yaml:"disabled"`
	HostIDs     []string          `json:"hostIds" yaml:"hostIds"`
	Memory      MemoryConfig      `json:"memory" yaml:"memory"`
	CPU         CPUConfig         `json:"cpu" yaml:"cpu"`
	Disk        DiskConfig        `json:"disk" yaml:"disk"`
	Temperature TemperatureConfig `json:"temperature" yaml:"temperature"`
	Process     ProcessConfig     `json:"process" yaml:"process"`
}

// Memory details
type MemoryConfig struct {
	Interval       string `json:"interval" yaml:"interval"`
	MemoryDisabled bool   `json:"memoryDisabled" yaml:"memoryDisabled"`
	SwapDisabled   bool   `json:"swapDisabled" yaml:"swapDisabled"`
	Unit           string `json:"unit" yaml:"unit"`
}

// CPU details
type CPUConfig struct {
	Interval       string `json:"interval" yaml:"interval"`
	CPUDisabled    bool   `json:"cpuDisabled" yaml:"cpuDisabled"`
	PerCPUDisabled bool   `json:"perCpuDisabled" yaml:"perCpuDisabled"`
}

// Disk details
type DiskConfig struct {
	Interval string              `json:"interval" yaml:"interval"`
	Disabled bool                `json:"disabled" yaml:"disabled"`
	Data     map[string]DiskData `json:"data" yaml:"data"`
}

// DiskData struct
type DiskData struct {
	Disabled bool   `json:"disabled" yaml:"disabled"`
	Name     string `json:"name" yaml:"name"`
	Path     string `json:"path" yaml:"path"`
	Unit     string `json:"unit" yaml:"unit"`
}

// Temperature details
type TemperatureConfig struct {
	Interval    string   `json:"interval" yaml:"interval"`
	DisabledAll bool     `json:"disabledAll" yaml:"disabledAll"`
	Enabled     []string `json:"enabled" yaml:"enabled"`
}

// Process details
type ProcessConfig struct {
	Interval string                 `json:"interval" yaml:"interval"`
	Disabled bool                   `json:"disabled" yaml:"disabled"`
	Data     map[string]ProcessData `json:"data" yaml:"data"`
}

// ProcessData struct
type ProcessData struct {
	Disabled bool              `json:"disabled" yaml:"disabled"`
	Name     string            `json:"name" yaml:"name"`
	Unit     string            `json:"unit" yaml:"unit"`
	Filter   map[string]string `json:"filter" yaml:"filter"`
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

const (
	// process filter fields
	ProcessFieldPid      = "pid"
	ProcessFieldCmdLine  = "cmdline"
	ProcessFieldCwd      = "cwd"
	ProcessFieldEXE      = "exe"
	ProcessFieldName     = "name"
	ProcessFieldNice     = "nice"
	ProcessFieldPPid     = "ppid"
	ProcessFieldUsername = "username"

	// process extra fields
	ProcessFieldGids          = "gids"
	ProcessFieldUids          = "uids"
	ProcessFieldCpuPercent    = "cpu_percent"
	ProcessFieldMemoryPercent = "memory_percent"
	ProcessFieldRSS           = "rss"
	ProcessFieldVMS           = "vms"
	ProcessFieldSwap          = "swap"
	ProcessFieldStack         = "stack"
	ProcessFieldLocked        = "locked"
	ProcessFieldData          = "data"
)
