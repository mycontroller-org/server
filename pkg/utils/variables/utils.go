package variables

import (
	"fmt"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	quickIdUL "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	tplUtils "github.com/mycontroller-org/backend/v2/pkg/utils/template"
	"go.uber.org/zap"
)

// LoadVariables loads all the defined variables
func LoadVariables(variablesPreMap map[string]string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	for name, quickID := range variablesPreMap {
		value := getEntity(name, quickID)
		if value == nil {
			return nil, fmt.Errorf("failed to load a variable. name: %s, selector:%s", name, quickID)
		}
		variables[name] = value
	}
	return variables, nil
}

func getEntity(name, quickID string) interface{} {
	if !quickIdUL.IsValidQuickID(quickID) {
		return quickID
	}

	resourceType, keys, err := quickIdUL.ResourceKeyValueMap(quickID)
	if err != nil {
		zap.L().Warn("failed to parse variable", zap.Any("name", name), zap.String("quickID", quickID), zap.Error(err))
		return nil
	}
	var entity interface{}
	switch {

	case utils.ContainsString(quickIdUL.QuickIDGateway, resourceType):
		item, err := gatewayAPI.GetByID(keys[model.KeyGatewayID])
		entity = item
		if err != nil {
			zap.L().Warn("gateway not available", zap.Any("keys", keys))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDNode, resourceType):
		item, err := nodeAPI.GetByGatewayAndNodeID(keys[model.KeyGatewayID], keys[model.KeyNodeID])
		entity = item
		if err != nil {
			zap.L().Warn("node not available", zap.Any("keys", keys))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDSensor, resourceType):
		item, err := sensorAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySensorID])
		entity = item
		if err != nil {
			zap.L().Warn("sensor not available", zap.Any("keys", keys))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDSensorField, resourceType):
		item, err := fieldAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySensorID], keys[model.KeyFieldID])
		entity = item
		if err != nil {
			zap.L().Warn("field not available", zap.Any("keys", keys), zap.Error(err))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDTemplate, resourceType):
		data, err := tplUtils.Execute(keys[model.KeyTemplate], nil)
		if err != nil {
			zap.L().Warn("failed to parse template", zap.Any("keys", keys), zap.Error(err))
			return nil
		}
		entity = data

	default:
		entity = nil
	}
	if entity == nil {
		return nil
	}

	if selector, ok := keys[model.KeySelector]; ok {
		_, value, err := helper.GetValueByKeyPath(entity, selector)
		if err != nil {
			zap.L().Warn("failed to selector", zap.Error(err), zap.String("selector", selector), zap.Any("entity", entity))
			return nil
		}
		return value
	}
	return entity
}
