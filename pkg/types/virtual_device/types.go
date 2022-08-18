package virtual_device

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

const (
	ResourceByQuickID = "resource_by_quick_id"
	ResourceByLabels  = "resource_by_labels"
)

// virtual device work is in progress
// This actually created to address Google Assistant home graph map
// Needs to be addressed lot of things
// for now this is incomplete and not usable

// VirtualDevice struct
type VirtualDevice struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	DeviceType  string               `json:"deviceType"`
	Traits      map[string]Resource  `json:"traits"`
	Location    string               `json:"location"`
	Labels      cmap.CustomStringMap `json:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn"`
	Resources   []string             `json:"resources"`
}

type Resource struct {
	Type           string               `json:"type"`
	ResourceType   string               `json:"resourceType"`
	QuickID        string               `json:"quickId"`
	Labels         cmap.CustomStringMap `json:"labels"`
	Value          interface{}          `json:"-"`
	ValueTimestamp time.Time            `json:"-"`
}
