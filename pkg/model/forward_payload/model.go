package forwardpayload

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Mapping of forward payload
type Mapping struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	SrcFieldID  string               `json:"srcFieldId"`
	DstFieldID  string               `json:"dstFieldId"`
	Labels      cmap.CustomStringMap `json:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn"`
}
