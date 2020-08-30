package node

import (
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Node functions
const (
	FuncReboot          = "reboot"
	FuncReset           = "reset"
	FuncDiscover        = "discover"
	FuncRefreshNodeInfo = "refresh_node_info"
	FuncHeartbeat       = "heartbeat"
)

// Node model
type Node struct {
	ID        string               `json:"id"`
	GatewayID string               `json:"gatewayId"`
	Name      string               `json:"name"`
	Labels    cmap.CustomStringMap `json:"labels"`
	Others    cmap.CustomMap       `json:"others"`
	Status    ml.State             `json:"status"`
	LastSeen  time.Time            `json:"lastSeen"`
}
