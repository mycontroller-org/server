package virtualdevice

import (
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// virtual device work is in progress
// This actually created to address Google Assistant home graph map
// Needs to be addressed lot of things
// for now this is incomplete and not usable

// VirtualDevice model
type VirtualDevice struct {
	ID          string               `json:"id"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	Labels      cmap.CustomStringMap `json:"labels"`
	DeviceType  string               `json:"deviceType"`
	Traits      cmap.CustomStringMap `json:"traits"`
	Parameters  cmap.CustomMap       `json:"parameters"`
}
