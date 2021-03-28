package source

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Source model
type Source struct {
	ID         string               `json:"id"`
	GatewayID  string               `json:"gatewayId"`
	NodeID     string               `json:"nodeId"`
	SourceID   string               `json:"sourceId"`
	Name       string               `json:"name"`
	Labels     cmap.CustomStringMap `json:"labels"`
	Others     cmap.CustomMap       `json:"others"`
	LastSeen   time.Time            `json:"lastSeen"`
	ModifiedOn time.Time            `json:"modifiedOn"`
}
