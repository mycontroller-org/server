package busutils

import (
	eventML "github.com/mycontroller-org/backend/v2/pkg/model/bus/event"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	filterHelper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	quickid "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	"go.uber.org/zap"
)

// PostEvent sends resource as event.
func PostEvent(eventTopic, eventType, entityType string, entity interface{}) {
	event := &eventML.Event{
		Type:       eventType,
		EntityType: entityType,
		Entity:     entity,
		EntityID:   filterHelper.GetID(entity),
	}

	quickID, _ := quickid.GetQuickID(entity)
	event.EntityQuickID = quickID

	err := mcbus.Publish(mcbus.FormatTopic(eventTopic), event)
	if err != nil {
		zap.L().Error("error on posting resource data", zap.String("topic", eventTopic), zap.Any("event", event), zap.Error(err))
	}
}
