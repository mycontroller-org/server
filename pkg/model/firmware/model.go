package firmware

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// reserved labels
const (
	LabelPlatform = "platform"
)

// Firmware struct
type Firmware struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Version    string               `json:"version"`
	File       FileConfig           `json:"file"`
	Labels     cmap.CustomStringMap `json:"labels"`
	ModifiedOn time.Time            `json:"modifiedOn"`
}

// FileConfig struct
type FileConfig struct {
	Name         string    `json:"name"`
	InternalName string    `json:"internalName"`
	Checksum     string    `json:"checksum"`
	Size         int       `json:"size"`
	ModifiedOn   time.Time `json:"modifiedOn"`
}
