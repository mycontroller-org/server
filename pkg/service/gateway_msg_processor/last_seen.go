package gatewaymessageprocessor

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"go.uber.org/zap"
)

// updates node last seen timestamp
func (svc *MessageProcessor) updateNodeLastSeen(gatewayID, nodeID string, timestamp time.Time) {
	node, err := svc.api.Node().GetByGatewayAndNodeID(gatewayID, nodeID)
	if err != nil {
		svc.logger.Debug("error on getting a node", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.Error(err))
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

	err = svc.api.Node().Save(node, true)
	if err != nil {
		svc.logger.Error("error on updating a node", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.Error(err))
	}

	// post node data to event listeners
	busUtils.PostEvent(svc.logger, svc.bus, topic.TopicEventNode, eventTY.TypeUpdated, types.EntityNode, node)
}

// updates source last seen timestamp
func (svc *MessageProcessor) updateSourceLastSeen(gatewayID, nodeID, sourceID string, timestamp time.Time) {
	source, err := svc.api.Source().GetByIDs(gatewayID, nodeID, sourceID)
	if err != nil {
		svc.logger.Debug("error on getting a source", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.String("sourceId", sourceID), zap.Error(err))
		return
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	// update lastseen
	source.LastSeen = timestamp

	err = svc.api.Source().Save(source)
	if err != nil {
		svc.logger.Debug("error on updating a source", zap.String("gatewayId", gatewayID), zap.String("nodeId", nodeID), zap.String("sourceId", sourceID), zap.Error(err))
	}

	// post source data to event listeners
	busUtils.PostEvent(svc.logger, svc.bus, topic.TopicEventSource, eventTY.TypeUpdated, types.EntityNode, source)
}
