package mysensors

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	topicTY "github.com/mycontroller-org/server/v2/pkg/types/topic"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

const (
	eventQueueLimit       = 50
	defaultEventQueueName = "event_listener_gw"
)

var (
	eventsQueue            *queueUtils.Queue
	firmwareSubscriptionID = int64(0)
	nodeSubscriptionID     = int64(0)
)

// initEventListener service
func (p *Provider) initEventListener(gatewayID string) error {
	firmwareEventQueueName := fmt.Sprintf("%s_%s", defaultEventQueueName, gatewayID)
	eventsQueue = queueUtils.New(p.logger, firmwareEventQueueName, eventQueueLimit, p.processServiceEvent, 1)
	// on message receive add it in to our local queue
	sID, err := p.bus.Subscribe(topicTY.TopicEventFirmware, p.onEvent)
	if err != nil {
		return err
	}
	firmwareSubscriptionID = sID
	sID, err = p.bus.Subscribe(topicTY.TopicEventNode, p.onEvent)
	if err != nil {
		return err
	}
	nodeSubscriptionID = sID
	return nil
}

// closeEventListener service
func (p *Provider) closeEventListener() {
	if firmwareSubscriptionID != 0 {
		topic := topicTY.TopicEventFirmware
		err := p.bus.Unsubscribe(topic, firmwareSubscriptionID)
		if err != nil {
			p.logger.Error("error on unsubscribe", zap.Error(err), zap.String("topic", topic))
		}
	}
	if nodeSubscriptionID != 0 {
		topic := topicTY.TopicEventNode
		err := p.bus.Unsubscribe(topic, nodeSubscriptionID)
		if err != nil {
			p.logger.Error("error on unsubscribe", zap.Error(err), zap.String("topic", topic))
		}
	}
	eventsQueue.Close()
}

func (p *Provider) onEvent(data *busTY.BusData) {
	event := &eventTY.Event{}
	err := data.LoadData(event)
	if err != nil {
		p.logger.Warn("Failed to convert to target type", zap.Error(err))
		return
	}
	p.logger.Debug("Received an event", zap.Any("event", event))

	if !(event.EntityType == types.EntityNode || event.EntityType == types.EntityFirmware) ||
		event.Entity == nil {
		return
	}

	p.logger.Debug("Event added into processing queue", zap.Any("event", event))
	status := eventsQueue.Produce(event)
	if !status {
		p.logger.Warn("Failed to store the event into queue", zap.Any("event", event))
	}
}

// processServiceEvent from the queue
func (p *Provider) processServiceEvent(item interface{}) {
	event := item.(*eventTY.Event)
	p.logger.Debug("Processing a request", zap.Any("event", event))

	// process events
	switch event.EntityType {
	case types.EntityFirmware:
		firmware := firmwareTY.Firmware{}
		err := event.LoadEntity(&firmware)
		if err != nil {
			p.logger.Error("error on loading firmware entity", zap.String("eventQuickId", event.EntityQuickID), zap.Error(err))
			return
		}
		fwRawStore.Remove(firmware.ID)
		fwStore.Remove(firmware.ID)

	case types.EntityNode:
		node := nodeTY.Node{}
		err := event.LoadEntity(&node)
		if err != nil {
			p.logger.Error("error on loading node entity", zap.String("eventQuickId", event.EntityQuickID), zap.Error(err))
			return
		}
		localID := p.getNodeStoreID(node.GatewayID, node.NodeID)
		if nodeStore.IsAvailable(localID) {
			nodeStore.Add(localID, &node)
		}

	default:
		p.logger.Info("received unsupported event", zap.Any("event", event))
	}
}
