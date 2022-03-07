package config

// Config data
type Config struct {
	HostConfigMap map[string]HostConfig `json:"hostConfigMap"`
}

type HostConfig struct {
	Disabled    bool              `json:"disabled"`
	HostIDs     []string          `json:"hostIds"`
	Memory      MemoryConfig      `json:"memory"`
	CPU         CPUConfig         `json:"cpu"`
	Disk        DiskConfig        `json:"disk"`
	Temperature TemperatureConfig `json:"temperature"`
	Process     ProcessConfig     `json:"process"`
}

// Memory details
type MemoryConfig struct {
	Interval       string `json:"interval"`
	MemoryDisabled bool   `json:"memoryDisabled"`
	SwapDisabled   bool   `json:"swapDisabled"`
	Unit           string `json:"unit"`
}

// CPU details
type CPUConfig struct {
	Interval       string `json:"interval"`
	CPUDisabled    bool   `json:"cpuDisabled"`
	PerCPUDisabled bool   `json:"perCpuDisabled"`
}

// Disk details
type DiskConfig struct {
	Interval string              `json:"interval"`
	Disabled bool                `json:"disabled"`
	Data     map[string]DiskData `json:"data"`
}

// DiskData struct
type DiskData struct {
	Disabled bool   `json:"disabled"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Unit     string `json:"unit"`
}

// Temperature details
type TemperatureConfig struct {
	Interval    string   `json:"interval"`
	DisabledAll bool     `json:"disabledAll"`
	Enabled     []string `json:"enabled"`
}

// Process details
type ProcessConfig struct {
	Interval string                 `json:"interval"`
	Disabled bool                   `json:"disabled"`
	Data     map[string]ProcessData `json:"data"`
}

// ProcessData struct
type ProcessData struct {
	Disabled bool              `json:"disabled"`
	Name     string            `json:"name"`
	Unit     string            `json:"unit"`
	Filter   map[string]string `json:"filter"`
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
