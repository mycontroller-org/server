package action

import (
	"fmt"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"go.uber.org/zap"
)

// ExecuteNodeAction for a node
func (a *ActionAPI) ExecuteNodeAction(action string, nodeIDs []string) error {
	// verify is a valid action?
	switch action {
	case nodeTY.ActionFirmwareUpdate,
		nodeTY.ActionHeartbeatRequest,
		nodeTY.ActionReboot,
		nodeTY.ActionRefreshNodeInfo,
		nodeTY.ActionReset:
		// nothing to do, just continue
	default:
		return fmt.Errorf("invalid node action:%s", action)
	}

	nodes, err := a.api.Node().GetByIDs(nodeIDs)
	if err != nil {
		return err
	}
	for index := 0; index < len(nodes); index++ {
		node := nodes[index]
		err = a.toNode(&node, node.GatewayID, node.NodeID, action)
		if err != nil {
			a.logger.Error("error on sending an action to a node", zap.Error(err), zap.String("gateway", node.GatewayID), zap.String("node", node.NodeID))
		}
	}
	return nil
}

func (a *ActionAPI) toNode(node *nodeTY.Node, gatewayID, nodeID, action string) error {
	msg := msgTY.NewMessage(false)
	msg.GatewayID = gatewayID
	msg.NodeID = nodeID

	// get node details and update isSleepNode
	if node == nil {
		node, err := a.api.Node().GetByGatewayAndNodeID(gatewayID, nodeID)
		if err == nil {
			msg.IsSleepNode = node.IsSleepNode()
		}
	} else {
		msg.IsSleepNode = node.IsSleepNode()
	}

	pl := msgTY.NewPayload()
	pl.Key = action
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgTY.TypeAction
	return a.Post(&msg)
}
