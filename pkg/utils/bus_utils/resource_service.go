package busutils

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"go.uber.org/zap"
)

// SetGatewayState send gateway status into bus
func SetGatewayState(id string, state types.State) {
	PostToResourceService(id, state, rsTY.TypeGateway, rsTY.CommandUpdateState, "")
}

// SetHandlerState send handler status into bus
func SetHandlerState(id string, state types.State) {
	PostToResourceService(id, state, rsTY.TypeHandler, rsTY.CommandUpdateState, "")
}

// SetTaskState send handler status into bus
func SetTaskState(id string, state taskTY.State) {
	PostToResourceService(id, state, rsTY.TypeTask, rsTY.CommandUpdateState, "")
}

// SetScheduleState send handler status into bus
func SetScheduleState(id string, state scheduleTY.State) {
	PostToResourceService(id, state, rsTY.TypeScheduler, rsTY.CommandUpdateState, "")
}

// DisableSchedule sends id to resource service
func DisableSchedule(id string) {
	PostToResourceService(id, id, rsTY.TypeScheduler, rsTY.CommandDisable, "")
}

// DisableTask sends id to resource service
func DisableTask(id string) {
	PostToResourceService(id, id, rsTY.TypeTask, rsTY.CommandDisable, "")
}

// PostToResourceService to resource service
func PostToResourceService(id string, data interface{}, serviceType, command, replyTopic string) {
	PostToService(mcbus.TopicServiceResourceServer, id, data, serviceType, command, replyTopic)
}

// PostToService posts to a service
func PostToService(sericeTopic, id string, data interface{}, serviceType, command, replyTopic string) {
	event := &rsTY.ServiceEvent{
		Type:       serviceType,
		Command:    command,
		ID:         id,
		ReplyTopic: replyTopic,
	}
	event.SetData(data)

	topic := mcbus.FormatTopic(sericeTopic)
	err := mcbus.Publish(topic, event)
	if err != nil {
		zap.L().Error("failed to post an event", zap.String("topic", topic), zap.Any("event", event))
	}
}
