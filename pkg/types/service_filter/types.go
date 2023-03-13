package servicefilter

import "github.com/mycontroller-org/server/v2/pkg/types/cmap"

// ServiceFilter struct
type ServiceFilter struct {
	Disabled bool                 `yaml:"disabled"`
	MatchAll bool                 `yaml:"match_all"`
	Types    []string             `yaml:"types"`
	IDs      []string             `yaml:"ids"`
	Labels   cmap.CustomStringMap `yaml:"labels"`
}

// HasFilter returns the filter status
func (sf *ServiceFilter) HasFilter() bool {
	return len(sf.Types) > 0 || len(sf.IDs) > 0 || len(sf.Labels) > 0
}

func (sf *ServiceFilter) Clone() *ServiceFilter {
	return &ServiceFilter{
		Disabled: sf.Disabled,
		MatchAll: sf.MatchAll,
		Types:    sf.Types,
		IDs:      sf.IDs,
		Labels:   sf.Labels.Clone(),
	}
}
