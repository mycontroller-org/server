package backup

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/version"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// global backup/restore constants
const (
	TypeJSON = "json"
	TypeYAML = "yaml"

	LimitPerFile         = 50
	EntityNameIndexSplit = "__"
	DateSuffixLayout     = "20060102_150405"

	DefaultStorageExporter = "default_storage_exporter"
	BackupDetailsFilename  = "backup.yaml"
)

// BackupDetails of a export
type BackupDetails struct {
	Filename          string          `yaml:"filename"`
	StorageExportType string          `yaml:"storage_export_type"`
	CreatedOn         time.Time       `yaml:"created_on"`
	Version           version.Version `yaml:"version"`
}

// BackupFile details
type BackupFile struct {
	ID           string    `json:"id"`
	LocationName string    `json:"locationName"`
	ProviderType string    `json:"providerType"`
	Directory    string    `json:"directory"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	FullPath     string    `json:"fullPath"`
	ModifiedOn   time.Time `json:"modifiedOn"`
}

// OnDemandBackupConfig config
type OnDemandBackupConfig struct {
	Prefix            string `json:"prefix"`
	StorageExportType string `json:"storageExportType"`
	TargetLocation    string `json:"targetLocation"`
	Handler           string `json:"handler"`
}

// BackupLocationDisk details
type BackupLocationDisk struct {
	TargetDirectory string
}

// holds save api details
type SaveAPIHolder struct {
	EntityType interface{}
	API        func(data interface{}) error
}

type ListFunc func(f []storageTY.Filter, p *storageTY.Pagination) (*storageTY.Result, error)
