package datarepository

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Config of data element
type Config struct {
	ID          string               `json:"id" yaml:"id"`
	ReadOnly    bool                 `json:"readOnly" yaml:"readOnly"`
	Description string               `json:"description" yaml:"description"`
	Labels      cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Data        cmap.CustomMap       `json:"data" yaml:"data"`
	ModifiedOn  time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}
