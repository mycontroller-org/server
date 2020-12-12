package export

// global export/import contants
const (
	TypeJSON = "json"
	TypeYAML = "yaml"

	ExporterNone = "none"

	LimitPerFile         = 50
	EntityNameIndexSplit = "__"

	DateSuffixLayout = "20060102_150405"
)

// Config for export job
type Config struct {
	Enabled       bool     `json:"enabled"`
	Interval      string   `json:"interval"`
	TargetDir     string   `json:"targetDir"`
	Clean         bool     `json:"clean"`
	UseDateSuffix bool     `json:"useDateSuffix"`
	ExportType    []string `json:"exportType"`
	Exporter      []string `json:"exporter"`
}

// ExporterConfig data
type ExporterConfig struct {
	Type    string                 `json:"type"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}
