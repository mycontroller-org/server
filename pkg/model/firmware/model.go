package firmware

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// reserved labels
const (
	LabelPlatform = "platform"
	BlockSize     = 512 // bytes
)

// Firmware struct
type Firmware struct {
	ID          string               `json:"id"`
	Description string               `json:"description"`
	File        FileConfig           `json:"file"`
	Labels      cmap.CustomStringMap `json:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn"`
}

// FileConfig struct
type FileConfig struct {
	Name         string    `json:"name"`
	InternalName string    `json:"internalName"`
	Checksum     string    `json:"checksum"`
	Size         int       `json:"size"`
	ModifiedOn   time.Time `json:"modifiedOn"`
}

type FirmwareBlock struct {
	ID    string `json:"id"`
	Block int64  `json:"block"`
	Total int64  `json:"total"`
	Data  []byte `json:"data"`
}
