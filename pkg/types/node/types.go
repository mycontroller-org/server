package node

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Node actions
const (
	ActionFirmwareUpdate   = "firmware_update"
	ActionHeartbeatRequest = "heartbeat_request"
	ActionReboot           = "reboot"
	ActionRefreshNodeInfo  = "refresh_node_info"
	ActionReset            = "reset"
	ActionAwake            = "awake"
)

// Node struct
type Node struct {
	ID         string               `json:"id"`
	GatewayID  string               `json:"gatewayId"`
	NodeID     string               `json:"nodeId"`
	Name       string               `json:"name"`
	Labels     cmap.CustomStringMap `json:"labels"`
	Others     cmap.CustomMap       `json:"others"`
	State      types.State          `json:"state"`
	LastSeen   time.Time            `json:"lastSeen"`
	ModifiedOn time.Time            `json:"modifiedOn"`
}
