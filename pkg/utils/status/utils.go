package status

import (
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

// SetGatewayState send gateway status into bus
func SetGatewayState(gatewayID string, state model.State) {
	event := &rsml.Event{
		Type:    rsml.TypeGateway,
		Command: rsml.CommandUpdateState,
		ID:      gatewayID,
	}
	event.SetData(state)
	topic := mcbus.FormatTopic(mcbus.TopicResourceServer)
	err := mcbus.Publish(topic, event)
	if err != nil {
		zap.L().Error("failed to post an event", zap.String("topic", topic), zap.Any("event", event))
	}
}
