package firmware

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// reserved labels
const (
	LabelPlatform = "platform"
	BlockSize     = 512 // bytes
)

// Firmware struct
type Firmware struct {
	ID          string               `json:"id" yaml:"id"`
	Description string               `json:"description" yaml:"description"`
	File        FileConfig           `json:"file" yaml:"file"`
	Labels      cmap.CustomStringMap `json:"labels" yaml:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}

// FileConfig struct
type FileConfig struct {
	Name         string    `json:"name" yaml:"name"`
	InternalName string    `json:"internalName" yaml:"internalName"`
	Checksum     string    `json:"checksum" yaml:"checksum"`
	Size         int       `json:"size" yaml:"size"`
	ModifiedOn   time.Time `json:"modifiedOn" yaml:"modifiedOn"`
}

type FirmwareBlock struct {
	ID          string `json:"id" yaml:"id"`
	BlockNumber int    `json:"blockNumber" yaml:"blockNumber"`
	TotalBytes  int    `json:"totalBytes" yaml:"totalBytes"` // entire file bytes size
	IsFinal     bool   `json:"isFinal" yaml:"isFinal"`
	Data        []byte `json:"data" yaml:"data"`
}
