package mysensors

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model"
	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	firmwareML "github.com/mycontroller-org/server/v2/pkg/model/firmware"
	nodeML "github.com/mycontroller-org/server/v2/pkg/model/node"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
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
func initEventListener(gatewayID string) error {
	firmwareEventQueueName := fmt.Sprintf("%s_%s", defaultEventQueueName, gatewayID)
	eventsQueue = queueUtils.New(firmwareEventQueueName, eventQueueLimit, processServiceEvent, 1)
	// on message receive add it in to our local queue
	sID, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicEventFirmware), onEvent)
	if err != nil {
		return err
	}
	firmwareSubscriptionID = sID
	sID, err = mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicEventNode), onEvent)
	if err != nil {
		return err
	}
	nodeSubscriptionID = sID
	return nil
}

// closeEventListener service
func closeEventListener() {
	if firmwareSubscriptionID != 0 {
		topic := mcbus.FormatTopic(mcbus.TopicEventFirmware)
		err := mcbus.Unsubscribe(topic, firmwareSubscriptionID)
		if err != nil {
			zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", topic))
		}
	}
	if nodeSubscriptionID != 0 {
		topic := mcbus.FormatTopic(mcbus.TopicEventNode)
		err := mcbus.Unsubscribe(topic, nodeSubscriptionID)
		if err != nil {
			zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", topic))
		}
	}
	eventsQueue.Close()
}

func onEvent(data *busML.BusData) {
	event := &eventML.Event{}
	err := data.LoadData(event)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}
	zap.L().Debug("Received an event", zap.Any("event", event))

	if !(event.EntityType == model.EntityNode || event.EntityType == model.EntityFirmware) ||
		event.Entity == nil {
		return
	}

	zap.L().Debug("Event added into processing queue", zap.Any("event", event))
	status := eventsQueue.Produce(event)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", event))
	}
}

// processServiceEvent from the queue
func processServiceEvent(item interface{}) {
	event := item.(*eventML.Event)
	zap.L().Debug("Processing a request", zap.Any("event", event))

	// process events
	if event.EntityType == model.EntityFirmware {
		if firmware, ok := event.Entity.(firmwareML.Firmware); ok {
			fwRawStore.Remove(firmware.ID)
			fwStore.Remove(firmware.ID)
		}
	} else if event.EntityType == model.EntityNode {
		if node, ok := event.Entity.(nodeML.Node); ok {
			localID := getNodeStoreID(node.GatewayID, node.NodeID)
			if nodeStore.IsAvailable(localID) {
				nodeStore.Add(localID, &node)
			}
		}
	}
}
