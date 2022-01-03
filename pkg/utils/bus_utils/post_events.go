package busutils

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	"go.uber.org/zap"
)

// PostEvent sends resource as event.
func PostEvent(eventTopic, eventType, entityType string, entity interface{}) {
	event := &eventTY.Event{
		Type:       eventType,
		EntityType: entityType,
		Entity:     entity,
		EntityID:   filterUtils.GetID(entity),
	}

	quickID, _ := quickIdUtils.GetQuickID(entity)
	event.EntityQuickID = quickID

	err := mcbus.Publish(mcbus.FormatTopic(eventTopic), event)
	if err != nil {
		zap.L().Error("error on posting resource data", zap.String("topic", eventTopic), zap.Any("event", event), zap.Error(err))
	}
}
