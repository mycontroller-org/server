package busutils

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	rsML "github.com/mycontroller-org/server/v2/pkg/model/resource_service"
	scheduleML "github.com/mycontroller-org/server/v2/pkg/model/schedule"
	taskML "github.com/mycontroller-org/server/v2/pkg/model/task"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"go.uber.org/zap"
)

// SetGatewayState send gateway status into bus
func SetGatewayState(id string, state model.State) {
	PostToResourceService(id, state, rsML.TypeGateway, rsML.CommandUpdateState, "")
}

// SetHandlerState send handler status into bus
func SetHandlerState(id string, state model.State) {
	PostToResourceService(id, state, rsML.TypeHandler, rsML.CommandUpdateState, "")
}

// SetTaskState send handler status into bus
func SetTaskState(id string, state taskML.State) {
	PostToResourceService(id, state, rsML.TypeTask, rsML.CommandUpdateState, "")
}

// SetScheduleState send handler status into bus
func SetScheduleState(id string, state scheduleML.State) {
	PostToResourceService(id, state, rsML.TypeScheduler, rsML.CommandUpdateState, "")
}

// DisableSchedule sends id to resource service
func DisableSchedule(id string) {
	PostToResourceService(id, id, rsML.TypeScheduler, rsML.CommandDisable, "")
}

// DisableTask sends id to resource service
func DisableTask(id string) {
	PostToResourceService(id, id, rsML.TypeTask, rsML.CommandDisable, "")
}

// PostToResourceService to resource service
func PostToResourceService(id string, data interface{}, serviceType, command, replyTopic string) {
	event := &rsML.ServiceEvent{
		Type:       serviceType,
		Command:    command,
		ID:         id,
		ReplyTopic: replyTopic,
	}
	event.SetData(data)

	topic := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	err := mcbus.Publish(topic, event)
	if err != nil {
		zap.L().Error("failed to post an event", zap.String("topic", topic), zap.Any("event", event))
	}
}
