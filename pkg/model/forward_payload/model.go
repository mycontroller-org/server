package forwardpayload

import "github.com/mycontroller-org/backend/v2/pkg/model/cmap"

// Mapping of forward payload
type Mapping struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	SourceID    string               `json:"sourceId"`
	TargetID    string               `json:"targetId"`
	Labels      cmap.CustomStringMap `json:"labels"`
}
