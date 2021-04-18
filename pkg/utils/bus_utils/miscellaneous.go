package busutils

import (
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"go.uber.org/zap"
)

// PostShutdownEvent asks the system to shutdown
func PostShutdownEvent() {
	err := mcbus.Publish(mcbus.FormatTopic(mcbus.TopicInternalShutdown), nil)
	if err != nil {
		zap.L().Error("error on posting shutdown event", zap.Error(err))
	}
}
