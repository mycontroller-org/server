package action

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	quickIdUL "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
)

type resourceAPI struct {
	Enable  func([]string) error
	Disable func([]string) error
	Reload  func([]string) error
}

func toResource(api resourceAPI, id, action string) error {
	switch action {
	case model.ActionEnable:
		return api.Enable([]string{id})

	case model.ActionDisable:
		return api.Disable([]string{id})

	case model.ActionReload:
		return api.Reload([]string{id})

	default:
		return fmt.Errorf("unknown action:%s", action)
	}

}

// ExecuteByQuickID the given request
func ExecuteByQuickID(quickID, payload string) error {
	resourceType, kvMap, err := quickIdUL.ResourceKeyValueMap(quickID)
	if err != nil {
		return err
	}

	switch {
	case utils.ContainsString(quickIdUL.QuickIDGateway, resourceType):
		return toGateway(kvMap[model.KeyGatewayID], payload)

	case utils.ContainsString(quickIdUL.QuickIDNode, resourceType):
		gatewayID := kvMap[model.KeyGatewayID]
		nodeID := kvMap[model.KeyNodeID]
		return toNode(gatewayID, nodeID, payload)

	case utils.ContainsString(quickIdUL.QuickIDSensor, resourceType):
		// no action needed

	case utils.ContainsString(quickIdUL.QuickIDSensorField, resourceType):
		return toSensorFieldByQuickID(kvMap, payload)

	case utils.ContainsString(quickIdUL.QuickIDTask, resourceType):
		return toTask(kvMap[model.KeyID], payload)

	case utils.ContainsString(quickIdUL.QuickIDSchedule, resourceType):
		return toSchedule(kvMap[model.KeyID], payload)

	default:
		return fmt.Errorf("Unknown resource type: %s", resourceType)
	}
	return nil
}

// Post a message to gateway topic
func Post(msg *msgml.Message) error {
	if msg.GatewayID == "" {
		return errors.New("gateway id can not be empty")
	}
	topic := mcbus.GetTopicPostMessageToProvider(msg.GatewayID)
	return mcbus.Publish(topic, msg)
}
