package forwardpayload

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Config of forward payload
type Config struct {
	ID          string               `json:"id" yaml:"id"`
	Description string               `json:"description" yaml:"description"`
	Enabled     bool                 `json:"enabled" yaml:"enabled"`
	SrcFieldID  string               `json:"srcFieldId" yaml:"srcFieldId"`
	DstFieldID  string               `json:"dstFieldId" yaml:"dstFieldId"`
	Labels      cmap.CustomStringMap `json:"labels" yaml:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}
