package gatewaymessageprocessor

import (
	"time"

	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	sourceAPI "github.com/mycontroller-org/server/v2/pkg/api/source"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"go.uber.org/zap"
)

// updates node last seen timestamp
func updateNodeLastSeen(gatewayID, nodeID string, timestamp time.Time) {
	node, err := nodeAPI.GetByGatewayAndNodeID(gatewayID, nodeID)
	if err != nil {
		zap.L().Debug("error on getting a node", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.Error(err))
		return
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	// update lastseen
	node.LastSeen = timestamp
	// update node status
	if node.State.Status != types.StatusUp {
		node.State = types.State{
			Status: types.StatusUp,
			Since:  timestamp,
		}
	}

	err = nodeAPI.Save(node)
	if err != nil {
		zap.L().Error("error on updating a node", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.Error(err))
	}

	// post node data to event listeners
	busUtils.PostEvent(mcbus.TopicEventNode, eventTY.TypeUpdated, types.EntityNode, node)
}

// updates source last seen timestamp
func updateSourceLastSeen(gatewayID, nodeID, sourceID string, timestamp time.Time) {
	source, err := sourceAPI.GetByIDs(gatewayID, nodeID, sourceID)
	if err != nil {
		zap.L().Debug("error on getting a source", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.String("sourceId", sourceID), zap.Error(err))
		return
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	// update lastseen
	source.LastSeen = timestamp

	err = sourceAPI.Save(source)
	if err != nil {
		zap.L().Debug("error on updating a source", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.String("sourceId", sourceID), zap.Error(err))
	}

	// post source data to event listeners
	busUtils.PostEvent(mcbus.TopicEventSource, eventTY.TypeUpdated, types.EntityNode, source)
}
