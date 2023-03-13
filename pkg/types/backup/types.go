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
	Filename          string          `json:"filename" yaml:"filename"`
	StorageExportType string          `json:"storage_export_type" yaml:"storage_export_type"`
	CreatedOn         time.Time       `json:"created_on" yaml:"created_on"`
	Version           version.Version `json:"version" yaml:"version"`
}

// BackupFile details
type BackupFile struct {
	ID           string    `json:"id" yaml:"id"`
	LocationName string    `json:"locationName" yaml:"locationName"`
	ProviderType string    `json:"providerType" yaml:"providerType"`
	Directory    string    `json:"directory" yaml:"directory"`
	FileName     string    `json:"fileName" yaml:"fileName"`
	FileSize     int64     `json:"fileSize" yaml:"fileSize"`
	FullPath     string    `json:"fullPath" yaml:"fullPath"`
	ModifiedOn   time.Time `json:"modifiedOn" yaml:"modifiedOn"`
}

// OnDemandBackupConfig config
type OnDemandBackupConfig struct {
	Prefix            string `json:"prefix" yaml:"prefix"`
	StorageExportType string `json:"storageExportType" yaml:"storageExportType"`
	TargetLocation    string `json:"targetLocation" yaml:"targetLocation"`
	Handler           string `json:"handler" yaml:"handler"`
}

// BackupLocationDisk details
type BackupLocationDisk struct {
	TargetDirectory string
}

// used to import and export data to database via existing api
type Backup interface {
	Import(data interface{}) error
	List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error)
	GetEntityInterface() interface{}
}
