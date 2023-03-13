package busutils

import (
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// PostShutdownEvent asks the system to shutdown
func PostShutdownEvent(logger *zap.Logger, bus busTY.Plugin) {
	err := bus.Publish(topic.TopicInternalShutdown, nil)
	if err != nil {
		logger.Error("error on posting shutdown event", zap.Error(err))
	}
}
