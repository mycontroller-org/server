package action

import (
	"errors"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

type resourceAPI interface {
	Enable([]string) error
	Disable([]string) error
	Reload([]string) error
}

func (a *ActionAPI) toEnableDisableReloadAction(api resourceAPI, id, action string) error {
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
func (a *ActionAPI) ExecuteActionOnResourceByQuickID(data *handlerTY.ResourceData) error {
	resourceType, kvMap, err := quickIdUtils.EntityKeyValueMap(data.QuickID)
	if err != nil {
		return err
	}

	switch resourceType {
	case quickIdUtils.QuickIdGateway:
		return a.toEnableDisableReloadAction(a.api.Gateway(), kvMap[types.KeyGatewayID], data.Payload)

	case quickIdUtils.QuickIdNode:
		gatewayID := kvMap[types.KeyGatewayID]
		nodeID := kvMap[types.KeyNodeID]
		return a.toNode(nil, gatewayID, nodeID, data.Payload)

	case quickIdUtils.QuickIdSource:
		// no action needed

	case quickIdUtils.QuickIdField:
		return a.toField(kvMap[types.KeyGatewayID], kvMap[types.KeyNodeID], kvMap[types.KeySourceID], kvMap[types.KeyFieldID], data.Payload)

	case quickIdUtils.QuickIdTask:
		return a.toEnableDisableReloadAction(a.api.Task(), kvMap[types.KeyID], data.Payload)

	case quickIdUtils.QuickIdSchedule:
		return a.toEnableDisableReloadAction(a.api.Schedule(), kvMap[types.KeyID], data.Payload)

	case quickIdUtils.QuickIdHandler:
		return a.toEnableDisableReloadAction(a.api.Handler(), kvMap[types.KeyID], data.Payload)

	case quickIdUtils.QuickIdDataRepository:
		return a.toDataRepository(kvMap[types.KeyID], data.KeyPath, data.Payload)

	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return nil
}

// ExecuteActionOnResourceByLabels the given request
func (a *ActionAPI) ExecuteActionOnResourceByLabels(data *handlerTY.ResourceData) error {
	if len(data.Labels) == 0 {
		return errors.New("empty labels not allowed")
	}
	filters := a.getFilterFromLabel(data.Labels)
	pagination := &storageTY.Pagination{Limit: 100}

	switch data.ResourceType {
	case quickIdUtils.QuickIdGateway:
		result, err := a.api.Gateway().List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]gatewayTY.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = a.toEnableDisableReloadAction(a.api.Gateway(), item.ID, data.Payload)
			if err != nil {
				a.logger.Error("error on sending data", zap.Error(err), zap.String("gatewayID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case quickIdUtils.QuickIdNode:
		result, err := a.api.Node().List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]nodeTY.Node)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = a.toNode(&item, item.GatewayID, item.NodeID, data.Payload)
			if err != nil {
				a.logger.Error("error on sending data", zap.Error(err), zap.String("nodeID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case quickIdUtils.QuickIdSource:
		// no action needed

	case quickIdUtils.QuickIdField:
		result, err := a.api.Field().List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]fieldTY.Field)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = a.toField(item.GatewayID, item.NodeID, item.SourceID, item.FieldID, data.Payload)
			if err != nil {
				a.logger.Error("error on sending data", zap.Error(err), zap.String("fieldID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case quickIdUtils.QuickIdTask:
		result, err := a.api.Task().List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]taskTY.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = a.toEnableDisableReloadAction(a.api.Task(), item.ID, data.Payload)
			if err != nil {
				a.logger.Error("error on sending data", zap.Error(err), zap.String("taskID", item.ID), zap.String("payload", data.Payload))
			}
		}

	case quickIdUtils.QuickIdSchedule:
		result, err := a.api.Schedule().List(filters, pagination)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return nil
		}
		items := result.Data.(*[]schedulerTY.Config)
		for index := 0; index < len(*items); index++ {
			item := (*items)[index]
			err = a.toEnableDisableReloadAction(a.api.Schedule(), item.ID, data.Payload)
			if err != nil {
				a.logger.Error("error on sending data", zap.Error(err), zap.String("scheduleID", item.ID), zap.String("payload", data.Payload))
			}
		}

	default:
		return fmt.Errorf("unknown resource type: %s", data.ResourceType)
	}
	return nil
}

// posts a message to a gateway provider
func (a *ActionAPI) Post(msg *msgTY.Message) error {
	if msg.GatewayID == "" {
		return errors.New("gateway id can not be empty")
	}
	// include node labels
	if msg.NodeID != "" {
		node, err := a.api.Node().GetByGatewayAndNodeID(msg.GatewayID, msg.NodeID)
		if err != nil {
			a.logger.Debug("error on getting node details", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
		} else {
			msg.Labels = msg.Labels.Init()
			msg.Labels.CopyFrom(node.Labels)
		}
	}
	return a.bus.Publish(fmt.Sprintf("%s.%s", topic.TopicPostMessageToProvider, msg.GatewayID), msg)
}

func (a *ActionAPI) getFilterFromLabel(labels cmap.CustomStringMap) []storageTY.Filter {
	filters := make([]storageTY.Filter, 0)
	for key, value := range labels {
		filters = append(filters, storageTY.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: storageTY.OperatorEqual, Value: value})
	}
	return filters
}
