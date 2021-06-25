package forwardpayload

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
)

// Config of forward payload
type Config struct {
	ID          string               `json:"id"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	SrcFieldID  string               `json:"srcFieldId"`
	DstFieldID  string               `json:"dstFieldId"`
	Labels      cmap.CustomStringMap `json:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn"`
}
