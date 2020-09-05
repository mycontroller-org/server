package node

import (
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Node functions
const (
	FuncDiscover         = "discover"
	FuncFirmwareUpdate   = "firmware_update"
	FuncHeartbeatRequest = "heartbeat_request"
	FuncReboot           = "reboot"
	FuncRefreshNodeInfo  = "refresh_node_info"
	FuncReset            = "reset"
)

// Known labels
const (
	LabelAssignedFirmware = "assigned_firmware" // id of the assigned firmware
)

// Node model
type Node struct {
	ID        string               `json:"id"`
	GatewayID string               `json:"gatewayId"`
	NodeID    string               `json:"nodeId"`
	Name      string               `json:"name"`
	Labels    cmap.CustomStringMap `json:"labels"`
	Others    cmap.CustomMap       `json:"others"`
	Status    ml.State             `json:"status"`
	LastSeen  time.Time            `json:"lastSeen"`
}
