package servicefilter

import "github.com/mycontroller-org/backend/v2/pkg/model/cmap"

// ServiceFilter struct
type ServiceFilter struct {
	Disabled bool                 `yaml:"disabled"`
	MatchAll bool                 `yaml:"match_all"`
	Types    []string             `yaml:"types"`
	IDs      []string             `yaml:"ids"`
	Labels   cmap.CustomStringMap `yaml:"lables"`
}

// HasFilter returns the filter status
func (sf *ServiceFilter) HasFilter() bool {
	return len(sf.Types) > 0 || len(sf.IDs) > 0 || len(sf.Labels) > 0
}
