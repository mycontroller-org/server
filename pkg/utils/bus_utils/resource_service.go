package busutils

import (
	"github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// SetGatewayState send gateway status into bus
func SetGatewayState(logger *zap.Logger, bus busTY.Plugin, id string, state types.State) {
	PostToResourceService(logger, bus, id, state, rsTY.TypeGateway, rsTY.CommandUpdateState, "")
}

// SetVirtualAssistantState send assistant status into bus
func SetVirtualAssistantState(logger *zap.Logger, bus busTY.Plugin, id string, state types.State) {
	PostToResourceService(logger, bus, id, state, rsTY.TypeVirtualAssistant, rsTY.CommandUpdateState, "")
}

// SetHandlerState send handler status into bus
func SetHandlerState(logger *zap.Logger, bus busTY.Plugin, id string, state types.State) {
	PostToResourceService(logger, bus, id, state, rsTY.TypeHandler, rsTY.CommandUpdateState, "")
}

// SetTaskState send handler status into bus
func SetTaskState(logger *zap.Logger, bus busTY.Plugin, id string, state taskTY.State) {
	PostToResourceService(logger, bus, id, state, rsTY.TypeTask, rsTY.CommandUpdateState, "")
}

// SetScheduleState send handler status into bus
func SetScheduleState(logger *zap.Logger, bus busTY.Plugin, id string, state schedulerTY.State) {
	PostToResourceService(logger, bus, id, state, rsTY.TypeScheduler, rsTY.CommandUpdateState, "")
}

// DisableSchedule sends id to resource service
func DisableSchedule(logger *zap.Logger, bus busTY.Plugin, id string) {
	PostToResourceService(logger, bus, id, id, rsTY.TypeScheduler, rsTY.CommandDisable, "")
}

// DisableTask sends id to resource service
func DisableTask(logger *zap.Logger, bus busTY.Plugin, id string) {
	PostToResourceService(logger, bus, id, id, rsTY.TypeTask, rsTY.CommandDisable, "")
}

// EnableTask sends id to resource service
func EnableTask(logger *zap.Logger, bus busTY.Plugin, id string) {
	PostToResourceService(logger, bus, id, id, rsTY.TypeTask, rsTY.CommandEnable, "")
}

// PostToResourceService to resource service
func PostToResourceService(logger *zap.Logger, bus busTY.Plugin, id string, data interface{}, serviceType, command, replyTopic string) {
	PostToService(logger, bus, topic.TopicServiceResourceServer, id, data, serviceType, command, replyTopic)
}

// PostToService posts to a service
func PostToService(logger *zap.Logger, bus busTY.Plugin, serviceTopic, id string, data interface{}, serviceType, command, replyTopic string) {
	event := &rsTY.ServiceEvent{
		Type:       serviceType,
		Command:    command,
		ID:         id,
		ReplyTopic: replyTopic,
	}
	event.SetData(data)

	err := bus.Publish(serviceTopic, event)
	if err != nil {
		logger.Error("failed to post an event", zap.String("topic", serviceTopic), zap.Any("event", event))
	}
}
