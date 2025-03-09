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
	ID          string               `json:"id" yaml:"id"`
	Name        string               `json:"name" yaml:"name"`
	Description string               `json:"description" yaml:"description"`
	Enabled     bool                 `json:"enabled" yaml:"enabled"`
	DeviceType  string               `json:"deviceType" yaml:"deviceType"`
	Traits      []Resource           `json:"traits" yaml:"traits"`
	Location    string               `json:"location" yaml:"location"`
	Labels      cmap.CustomStringMap `json:"labels" yaml:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
	Resources   []string             `json:"resources" yaml:"resources"`
}

type Resource struct {
	Name           string               `json:"name" yaml:"name"`
	TraitType      string               `json:"traitType" yaml:"traitType"`
	ResourceType   string               `json:"resourceType"`
	QuickID        string               `json:"quickId" yaml:"quickId"`
	Labels         cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Value          interface{}          `json:"-" yaml:"-"`
	ValueTimestamp time.Time            `json:"-" yaml:"-"`
}
