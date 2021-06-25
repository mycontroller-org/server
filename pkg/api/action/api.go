package action

import (
	"errors"
	"fmt"

	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	fieldML "github.com/mycontroller-org/server/v2/pkg/model/field"
	gatewayML "github.com/mycontroller-org/server/v2/pkg/model/gateway"
	handlerML "github.com/mycontroller-org/server/v2/pkg/model/handler"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/server/v2/pkg/model/node"
	scheduleML "github.com/mycontroller-org/server/v2/pkg/model/schedule"
	taskML "github.com/mycontroller-org/server/v2/pkg/model/task"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	quickIdUL "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	stgML "github.com/mycontroller-org/server/v2/plugin/storage"
	"go.uber.org/zap"
)

type resourceAPI struct {
	Enable  func([]string) error
	Disable func([]string) error
	Reload  func([]string) error
}

func toResource(api resourceAPI, id, action string) error {
	action = model.GetAction(action)
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

// ExecuteActionOnResourceByQuickID the given request
func ExecuteActionOnResourceByQuickID(data *handlerML.ResourceData) error {
	resourceType, kvMap, err := quickIdUL.EntityKeyValueMap(data.QuickID)
	if err != nil {
		return err
	}

	switch {
	case utils.ContainsString(quickIdUL.QuickIDGateway, resourceType):
		return toGateway(kvMap[model.KeyGatewayID], data.Payload)

	case utils.ContainsString(quickIdUL.QuickIDNode, resourceType):
		gatewayID := kvMap[model.KeyGatewayID]
		nodeID := kvMap[model.KeyNodeID]
		return toNode(gatewayID, nodeID, data.Payload)

	case utils.ContainsString(quickIdUL.QuickIDSource, resourceType):
		// no action needed

	case utils.ContainsString(quickIdUL.QuickIDField, resourceType):
		return ToField(kvMap[model.KeyGatewayID], kvMap[model.KeyNodeID], kvMap[model.KeySourceID], kvMap[model.KeyFieldID], data.Payload)

	case utils.ContainsString(quickIdUL.QuickIDTask, resourceType):
		return toTask(kvMap[model.KeyID], data.Payload)

	case utils.ContainsString(quickIdUL.QuickIDSchedule, resourceType):
		return toSchedule(kvMap[model.KeyID], data.Payload)

	case utils.ContainsString(quickIdUL.QuickIDHandler, resourceType):
		return toHandler(kvMap[model.KeyID], data.Payload)

	case utils.ContainsString(quickIdUL.QuickIDDataRepository, resourceType):
		return toDataRepository(kvMap[model.KeyID], data.KeyPath, data.Payload)

	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return nil
}

// ExecuteActionOnResourceByLabels the given request
func ExecuteActionOnResourceByLabels(data *handlerML.ResourceData) error {
	if len(data.Labels) == 0 {
		return errors.New("empty labels not allowed")
	}
	filters := getFilterFromLabel(data.Labels)
	pagination := &stgML.Pagination{Limit: 100}

	switch {
	case utils.ContainsString(quickIdUL.QuickIDGateway, data.ResourceType):
		result, err := gatewayAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]gatewayML.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toGateway(item.ID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("gatewayID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDNode, data.ResourceType):
		result, err := nodeAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]nodeML.Node)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toNode(item.GatewayID, item.NodeID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("nodeID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDSource, data.ResourceType):
		// no action needed

	case utils.ContainsString(quickIdUL.QuickIDField, data.ResourceType):
		result, err := fieldAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]fieldML.Field)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = ToField(item.GatewayID, item.NodeID, item.SourceID, item.FieldID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("fieldID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDTask, data.ResourceType):
		result, err := taskAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]taskML.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toTask(item.ID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("taskID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDSchedule, data.ResourceType):
		result, err := scheduleAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]scheduleML.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toSchedule(item.ID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("scheduleID", item.ID), zap.String("payload", data.Payload))
			}
		}

	default:
		return fmt.Errorf("unknown resource type: %s", data.ResourceType)
	}
	return nil
}

// Post a message to gateway topic
func Post(msg *msgML.Message) error {
	if msg.GatewayID == "" {
		return errors.New("gateway id can not be empty")
	}
	topic := mcbus.GetTopicPostMessageToProvider(msg.GatewayID)
	return mcbus.Publish(topic, msg)
}

func getFilterFromLabel(labels cmap.CustomStringMap) []stgML.Filter {
	filters := make([]stgML.Filter, 0)
	for key, value := range labels {
		filters = append(filters, stgML.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: stgML.OperatorEqual, Value: value})
	}
	return filters
}
