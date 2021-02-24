package action

import (
	"errors"
	"fmt"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	gatewayML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	quickIdUL "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
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

// ExecuteActionOnResourceByQuickID the given request
func ExecuteActionOnResourceByQuickID(quickID, payload string) error {
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
		return toSensorField(kvMap[model.KeyGatewayID], kvMap[model.KeyNodeID], kvMap[model.KeySensorID], kvMap[model.KeyFieldID], payload)

	case utils.ContainsString(quickIdUL.QuickIDTask, resourceType):
		return toTask(kvMap[model.KeyID], payload)

	case utils.ContainsString(quickIdUL.QuickIDSchedule, resourceType):
		return toSchedule(kvMap[model.KeyID], payload)

	default:
		return fmt.Errorf("Unknown resource type: %s", resourceType)
	}
	return nil
}

// ExecuteActionOnResourceByLabels the given request
func ExecuteActionOnResourceByLabels(resourceType string, labels cmap.CustomStringMap, payload string) error {
	if len(labels) == 0 {
		return errors.New("empty labels not allowed")
	}
	filters := getFilterFromLabel(labels)
	pagination := &stgML.Pagination{Limit: 100}

	switch {
	case utils.ContainsString(quickIdUL.QuickIDGateway, resourceType):
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
			err = toGateway(item.ID, payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("gatewayID", item.ID), zap.String("payload", payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDNode, resourceType):
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
			err = toNode(item.GatewayID, item.NodeID, payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("nodeID", item.ID), zap.String("payload", payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDSensor, resourceType):
		// no action needed

	case utils.ContainsString(quickIdUL.QuickIDSensorField, resourceType):
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
			err = toSensorField(item.GatewayID, item.NodeID, item.SensorID, item.FieldID, payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("fieldID", item.ID), zap.String("payload", payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDTask, resourceType):
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
			err = toTask(item.ID, payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("taskID", item.ID), zap.String("payload", payload))
			}
		}

	case utils.ContainsString(quickIdUL.QuickIDSchedule, resourceType):
		result, err := schedulerAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]schedulerML.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toSchedule(item.ID, payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("scheduleID", item.ID), zap.String("payload", payload))
			}
		}

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

func getFilterFromLabel(labels cmap.CustomStringMap) []stgML.Filter {
	filters := make([]stgML.Filter, 0)
	for key, value := range labels {
		filters = append(filters, stgML.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: stgML.OperatorEqual, Value: value})
	}
	return filters
}
