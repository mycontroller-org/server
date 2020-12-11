package node

import (
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Node functions
const (
	ActionDiscover         = "discover"
	ActionFirmwareUpdate   = "firmware_update"
	ActionHeartbeatRequest = "heartbeat_request"
	ActionReboot           = "reboot"
	ActionRefreshNodeInfo  = "refresh_node_info"
	ActionReset            = "reset"
)

// Node model
type Node struct {
	ID             string               `json:"id"`
	GatewayID      string               `json:"gatewayId"`
	NodeID         string               `json:"nodeId"`
	Name           string               `json:"name"`
	Labels         cmap.CustomStringMap `json:"labels"`
	Others         cmap.CustomMap       `json:"others"`
	State          ml.State             `json:"state"`
	LastSeen       time.Time            `json:"lastSeen"`
	LastModifiedOn time.Time            `json:"lastModifiedOn"`
}
