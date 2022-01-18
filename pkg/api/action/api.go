package action

import (
	"errors"
	"fmt"

	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

type resourceAPI struct {
	Enable  func([]string) error
	Disable func([]string) error
	Reload  func([]string) error
}

func toResource(api resourceAPI, id, action string) error {
	action = types.GetAction(action)
	switch action {
	case types.ActionEnable:
		return api.Enable([]string{id})

	case types.ActionDisable:
		return api.Disable([]string{id})

	case types.ActionReload:
		return api.Reload([]string{id})

	default:
		return fmt.Errorf("unknown action:%s", action)
	}

}

// ExecuteActionOnResourceByQuickID the given request
func ExecuteActionOnResourceByQuickID(data *handlerTY.ResourceData) error {
	resourceType, kvMap, err := quickIdUtils.EntityKeyValueMap(data.QuickID)
	if err != nil {
		return err
	}

	switch {
	case utils.ContainsString(quickIdUtils.QuickIDGateway, resourceType):
		return toGateway(kvMap[types.KeyGatewayID], data.Payload)

	case utils.ContainsString(quickIdUtils.QuickIDNode, resourceType):
		gatewayID := kvMap[types.KeyGatewayID]
		nodeID := kvMap[types.KeyNodeID]
		return toNode(nil, gatewayID, nodeID, data.Payload)

	case utils.ContainsString(quickIdUtils.QuickIDSource, resourceType):
		// no action needed

	case utils.ContainsString(quickIdUtils.QuickIDField, resourceType):
		return toField(kvMap[types.KeyGatewayID], kvMap[types.KeyNodeID], kvMap[types.KeySourceID], kvMap[types.KeyFieldID], data.Payload)

	case utils.ContainsString(quickIdUtils.QuickIDTask, resourceType):
		return toTask(kvMap[types.KeyID], data.Payload)

	case utils.ContainsString(quickIdUtils.QuickIDSchedule, resourceType):
		return toSchedule(kvMap[types.KeyID], data.Payload)

	case utils.ContainsString(quickIdUtils.QuickIDHandler, resourceType):
		return toHandler(kvMap[types.KeyID], data.Payload)

	case utils.ContainsString(quickIdUtils.QuickIDDataRepository, resourceType):
		return toDataRepository(kvMap[types.KeyID], data.KeyPath, data.Payload)

	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return nil
}

// ExecuteActionOnResourceByLabels the given request
func ExecuteActionOnResourceByLabels(data *handlerTY.ResourceData) error {
	if len(data.Labels) == 0 {
		return errors.New("empty labels not allowed")
	}
	filters := getFilterFromLabel(data.Labels)
	pagination := &storageTY.Pagination{Limit: 100}

	switch {
	case utils.ContainsString(quickIdUtils.QuickIDGateway, data.ResourceType):
		result, err := gatewayAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]gatewayTY.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toGateway(item.ID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("gatewayID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUtils.QuickIDNode, data.ResourceType):
		result, err := nodeAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]nodeTY.Node)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toNode(&item, item.GatewayID, item.NodeID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("nodeID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUtils.QuickIDSource, data.ResourceType):
		// no action needed

	case utils.ContainsString(quickIdUtils.QuickIDField, data.ResourceType):
		result, err := fieldAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]fieldTY.Field)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toField(item.GatewayID, item.NodeID, item.SourceID, item.FieldID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("fieldID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUtils.QuickIDTask, data.ResourceType):
		result, err := taskAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]taskTY.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = toTask(item.ID, data.Payload)
			if err != nil {
				zap.L().Error("error on sending data", zap.Error(err), zap.String("taskID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case utils.ContainsString(quickIdUtils.QuickIDSchedule, data.ResourceType):
		result, err := scheduleAPI.List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]scheduleTY.Config)
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
func Post(msg *msgTY.Message) error {
	if msg.GatewayID == "" {
		return errors.New("gateway id can not be empty")
	}
	topic := mcbus.GetTopicPostMessageToProvider(msg.GatewayID)
	return mcbus.Publish(topic, msg)
}

func getFilterFromLabel(labels cmap.CustomStringMap) []storageTY.Filter {
	filters := make([]storageTY.Filter, 0)
	for key, value := range labels {
		filters = append(filters, storageTY.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: storageTY.OperatorEqual, Value: value})
	}
	return filters
}
