package helper

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"go.uber.org/zap"
)

const (
	JobTypeNodeStateUpdater = "node_state_updater"
	JobTypeSunriseUpdater   = "sunrise_updater"
)

// PostEvent sends job change notification.
func PostEvent(jobType string) {
	event := &rsTY.ServiceEvent{
		Type:    rsTY.TypeSystemJobs,
		Command: rsTY.CommandReload,
		Data:    jobType,
	}
	topic := mcbus.FormatTopic(mcbus.TopicInternalSystemJobs)
	err := mcbus.Publish(topic, event)
	if err != nil {
		zap.L().Error("error on posting resource data", zap.String("topic", topic), zap.Any("event", event), zap.Error(err))
	}
}
