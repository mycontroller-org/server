package gateway

import (
	"fmt"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/bus_utils/query"
)

const (
	queryTimeout = time.Second * 2
)

// Limitations: if multiple gateway sends data for a request
// for now shows only the first received data from a gateway

// returns a sleeping queue from a gateway
func (gw *GatewayAPI) GetGatewaySleepingQueue(gatewayID string) (map[string][]msgTY.Message, error) {
	ids := map[string]interface{}{
		types.KeyGatewayID: gatewayID,
	}

	messages := make(map[string][]msgTY.Message)
	onReceive := func(item interface{}) bool { return false }

	err := query.QueryService(gw.logger, gw.bus, topic.TopicServiceGateway, "", rsTY.TypeGateway, rsTY.CommandGetSleepingQueue, ids, onReceive, &messages, queryTimeout)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// returns a sleeping queue from a node
func (gw *GatewayAPI) GetNodeSleepingQueue(gatewayID, nodeID string) ([]msgTY.Message, error) {
	if gatewayID == "" || nodeID == "" {
		return nil, fmt.Errorf("gatewayId[%s] or nodeId[%s] can not be empty", gatewayID, nodeID)
	}
	ids := map[string]interface{}{
		types.KeyGatewayID: gatewayID,
		types.KeyNodeID:    nodeID,
	}

	messages := make([]msgTY.Message, 0)
	onReceive := func(item interface{}) bool { return false }

	err := query.QueryService(gw.logger, gw.bus, topic.TopicServiceGateway, "", rsTY.TypeGateway, rsTY.CommandGetSleepingQueue, ids, onReceive, &messages, queryTimeout)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (gw *GatewayAPI) ClearSleepingQueue(gatewayID, nodeID string) error {
	if gatewayID == "" {
		return fmt.Errorf("gatewayId[%s] can not be empty", gatewayID)
	}
	ids := map[string]interface{}{
		types.KeyGatewayID: gatewayID,
		types.KeyNodeID:    nodeID,
	}

	busUtils.PostToService(gw.logger, gw.bus, topic.TopicServiceGateway, "", ids, rsTY.TypeGateway, rsTY.CommandClearSleepingQueue, "")
	return nil
}
