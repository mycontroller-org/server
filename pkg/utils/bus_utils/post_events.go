package busutils

import (
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// PostEvent sends resource as event.
func PostEvent(logger *zap.Logger, bus busTY.Plugin, eventTopic, eventType, entityType string, entity interface{}) {
	event := &eventTY.Event{
		Type:       eventType,
		EntityType: entityType,
		Entity:     entity,
		EntityID:   filterUtils.GetID(entity),
	}

	quickID, _ := quickIdUtils.GetQuickID(entity)
	event.EntityQuickID = quickID

	err := bus.Publish(eventTopic, event)
	if err != nil {
		logger.Error("error on posting data", zap.String("topic", eventTopic), zap.Any("event", event), zap.Error(err))
	}
}

// PostEvent sends job change notification.
func PostServiceEvent(logger *zap.Logger, bus busTY.Plugin, topic, serviceType, serviceCommand, data string) {
	event := &rsTY.ServiceEvent{
		Type:    serviceType,
		Command: serviceCommand,
		Data:    data,
	}
	err := bus.Publish(topic, event)
	if err != nil {
		logger.Error("error on posting data", zap.String("topic", topic), zap.Any("event", event), zap.Error(err))
	}
}
