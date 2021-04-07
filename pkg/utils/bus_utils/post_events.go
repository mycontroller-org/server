package busutils

import (
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"go.uber.org/zap"
)

// PostEvent sends resource as event.
func PostEvent(eventTopic string, resource interface{}) {
	err := mcbus.Publish(mcbus.FormatTopic(eventTopic), resource)
	if err != nil {
		zap.L().Error("error on posting resource data", zap.String("topic", eventTopic), zap.Any("resource", resource), zap.Error(err))
	}
}
