package settings

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// key to get a specific settings
const (
	KeySystemSettings        = "system_settings"
	KeySystemJobs            = "system_jobs"
	KeySystemBackupLocations = "system_backup_locations"
	KeyVersion               = "version"
)

// Settings struct
type Settings struct {
	ID         string                 `json:"id"`
	Spec       map[string]interface{} `json:"spec"`
	ModifiedOn time.Time              `json:"modifiedOn"`
}

// SystemSettings struct
type SystemSettings struct {
	GeoLocation GeoLocation `json:"geoLocation"`
	Login       Login       `json:"login"`
}

// GeoLocation struct
type GeoLocation struct {
	AutoUpdate   bool    `json:"autoUpdate"`
	LocationName string  `json:"locationName"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// Login settings
type Login struct {
	Message string `json:"message"`
}

// VersionSettings struct
type VersionSettings struct {
	Version string `json:"version"`
}

// SystemJobsSettings cron struct
type SystemJobsSettings struct {
	Sunrise string `json:"sunrise"` // updates scheduled sunrise sunset jobs on this time
}

// BackupLocations of server
type BackupLocations struct {
	Locations []BackupLocation `json:"locations"`
}

// BackupLocation of server
type BackupLocation struct {
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Config cmap.CustomMap `json:"config"`
}
