package datarepository

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Config of data element
type Config struct {
	ID          string               `json:"id"`
	ReadOnly    bool                 `json:"readOnly"`
	Description string               `json:"description"`
	Labels      cmap.CustomStringMap `json:"labels"`
	Data        cmap.CustomMap       `json:"data"`
	ModifiedOn  time.Time            `json:"modifiedOn"`
}
