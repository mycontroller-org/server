package source

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Source struct
type Source struct {
	ID         string               `json:"id" yaml:"id"`
	GatewayID  string               `json:"gatewayId" yaml:"gatewayId"`
	NodeID     string               `json:"nodeId" yaml:"nodeId"`
	SourceID   string               `json:"sourceId" yaml:"sourceId"`
	Name       string               `json:"name" yaml:"name"`
	Labels     cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Others     cmap.CustomMap       `json:"others" yaml:"others"`
	LastSeen   time.Time            `json:"lastSeen" yaml:"lastSeen"`
	ModifiedOn time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}
