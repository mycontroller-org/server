package action

import (
	"fmt"

	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	"go.uber.org/zap"
)

// ExecuteNodeAction for a node
func ExecuteNodeAction(action string, nodeIDs []string) error {
	// verify is a valid action?
	switch action {
	case nodeML.ActionFirmwareUpdate,
		nodeML.ActionHeartbeatRequest,
		nodeML.ActionReboot,
		nodeML.ActionRefreshNodeInfo,
		nodeML.ActionReset:
		// nothing to do, just continue
	default:
		return fmt.Errorf("invalid node action:%s", action)
	}

	nodes, err := nodeAPI.GetByIDs(nodeIDs)
	if err != nil {
		return err
	}
	for index := 0; index < len(nodes); index++ {
		node := nodes[index]
		err = toNode(node.GatewayID, node.NodeID, action)
		if err != nil {
			zap.L().Error("error on sending an action to a node", zap.Error(err), zap.String("gateway", node.GatewayID), zap.String("node", node.NodeID))
		}
	}
	return nil
}

func toNode(gatewayID, nodeID, action string) error {
	msg := msgML.NewMessage(false)
	msg.GatewayID = gatewayID
	msg.NodeID = nodeID
	pl := msgML.NewPayload()
	pl.Key = action
	pl.Value = ""
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgML.TypeAction
	return Post(&msg)
}
