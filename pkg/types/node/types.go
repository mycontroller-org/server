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
	ID         string               `json:"id" yaml:"id"`
	GatewayID  string               `json:"gatewayId" yaml:"gatewayId"`
	NodeID     string               `json:"nodeId" yaml:"nodeId"`
	Name       string               `json:"name" yaml:"name"`
	Labels     cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Others     cmap.CustomMap       `json:"others" yaml:"others"`
	State      types.State          `json:"state" yaml:"state"`
	LastSeen   time.Time            `json:"lastSeen" yaml:"lastSeen"`
	ModifiedOn time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}

func (n *Node) IsSleepNode() bool {
	if n.Labels != nil && n.Labels.GetBool(types.LabelNodeSleepNode) {
		return true
	}
	return false
}
