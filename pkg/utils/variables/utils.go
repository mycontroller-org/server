package variables

import (
	"fmt"
	"reflect"
	"strings"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/notify_handler"
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	quickIdUL "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	tplUtils "github.com/mycontroller-org/backend/v2/pkg/utils/template"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

type genericAPI struct {
	List func(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error)
}

// LoadVariables loads all the defined variables
func LoadVariables(variablesPreMap map[string]string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	for name, stringValue := range variablesPreMap {
		value := getEntity(name, stringValue)
		if value == nil {
			return nil, fmt.Errorf("failed to load a variable. name: %s, selector:%s", name, stringValue)
		}
		variables[name] = value
	}
	return variables, nil
}

func getEntity(name, stringValue string) interface{} {
	rsData, err := GetResourceSelector(stringValue)
	if err != nil {
		return stringValue
	}

	if rsData.QuickID != "" {
		return getByQuickID(name, rsData)
	} else if rsData.ResourceType != "" && len(rsData.Labels) > 0 {
		return getByLabels(name, rsData)
	}

	return stringValue
}

func getByLabels(name string, rsData *rsML.ResourceSelector) interface{} {

	apiImpl := genericAPI{}
	switch {
	case utils.ContainsString(quickIdUL.QuickIDGateway, rsData.ResourceType):
		apiImpl.List = gatewayAPI.List

	case utils.ContainsString(quickIdUL.QuickIDNode, rsData.ResourceType):
		apiImpl.List = nodeAPI.List

	case utils.ContainsString(quickIdUL.QuickIDSensor, rsData.ResourceType):
		apiImpl.List = sensorAPI.List

	case utils.ContainsString(quickIdUL.QuickIDSensorField, rsData.ResourceType):
		apiImpl.List = fieldAPI.List

	case utils.ContainsString(quickIdUL.QuickIDTask, rsData.ResourceType):
		apiImpl.List = taskAPI.List

	case utils.ContainsString(quickIdUL.QuickIDSchedule, rsData.ResourceType):
		apiImpl.List = schedulerAPI.List

	case utils.ContainsString(quickIdUL.QuickIDHandler, rsData.ResourceType):
		apiImpl.List = handlerAPI.List
	}

	if apiImpl.List != nil {
		filters := make([]stgml.Filter, 0)
		for key, value := range rsData.Labels {
			filter := stgml.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: stgml.OperatorEqual, Value: value}
			filters = append(filters, filter)
		}
		pagination := &stgml.Pagination{Limit: 1} // limit to one element
		result, err := apiImpl.List(filters, pagination)
		if err != nil {
			return err.Error()
		}
		if result.Count == 0 {
			return nil
		}

		//get the first element
		if reflect.TypeOf(result.Data).Kind() == reflect.Slice {
			s := reflect.ValueOf(result.Data)
			if s.Len() == 0 {
				return nil
			}
			entity := s.Index(0)
			if rsData.Selector != "" {
				_, value, err := helper.GetValueByKeyPath(entity, rsData.Selector)
				if err != nil {
					zap.L().Warn("error on getting data from a given selector", zap.Error(err), zap.String("selector", rsData.Selector), zap.Any("entity", entity))
					return nil
				}
				return value
			}
			return entity
		}
	}

	return nil
}

func getByQuickID(name string, rsData *rsML.ResourceSelector) interface{} {
	resourceType, keys, err := quickIdUL.ResourceKeyValueMap(rsData.QuickID)
	if err != nil {
		zap.L().Warn("failed to parse variable", zap.Any("name", name), zap.Any("selector", rsData), zap.Error(err))
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

	default:
		data, err := tplUtils.Execute(keys[model.KeyTemplate], nil)
		if err != nil {
			zap.L().Warn("failed to parse template", zap.Any("keys", keys), zap.Error(err))
			return nil
		}
		entity = data
	}

	if entity == nil {
		return nil
	}

	if rsData.Selector != "" {
		zap.L().Debug("data", zap.Any("name", name), zap.Any("data", rsData), zap.Any("entity", entity))
		_, value, err := helper.GetValueByKeyPath(entity, rsData.Selector)
		if err != nil {
			zap.L().Warn("error on getting data from a given selector", zap.Error(err), zap.String("selector", rsData.Selector), zap.Any("entity", entity))
			return nil
		}
		return value
	}
	return entity
}

// Merge variables and parameters
func Merge(variables map[string]interface{}, parameters map[string]string) map[string]interface{} {
	finalMap := make(map[string]interface{})
	for key, value := range parameters { // update variables
		finalMap[key] = value
	}

	for key, value := range parameters { // execute as template
		updatedValue, err := tplUtils.Execute(value, variables)
		if err != nil {
			finalMap[key] = err.Error()
			continue
		}
		finalMap[key] = updatedValue
	}
	return finalMap
}

// UpdateParameters updates parmaeter templates
func UpdateParameters(variables map[string]interface{}, parameters map[string]string) {
	for name, value := range parameters {
		updatedValue, err := tplUtils.Execute(value, variables)
		if err != nil {
			parameters[name] = err.Error()
			continue
		}
		parameters[name] = updatedValue
	}
}

// GetAsMap returns resource string as map
// example parameters:
// --labels=true,--type=sf,--payload=increment|1|100|0,label=value1,label2=value2
// --qid=gw:mysensor,--payload=reload
// --qid=sf:mysensor.1.21.V_STATUS,--payload=on,--preDelay=1s
// --qid=sf:mysensor.1.21.V_STATUS,--selector=payload.value
func GetAsMap(stringData string) map[string]string {
	// replace ignored comma with different char
	innerCommaRetainer := "@#$@"
	newLabelString := strings.ReplaceAll(stringData, `\,`, innerCommaRetainer)
	labels := strings.Split(newLabelString, ",")

	data := make(map[string]string)
	for _, tmpLabel := range labels {
		label := strings.ReplaceAll(tmpLabel, innerCommaRetainer, ",")
		keyValue := strings.SplitN(label, "=", 2)
		key := strings.ToLower(strings.TrimSpace(keyValue[0]))
		value := ""
		if key == "" {
			continue
		}
		if len(keyValue) == 2 {
			value = strings.TrimSpace(keyValue[1])
		}
		data[key] = value
	}
	return data
}

// GetResourceSelector returns resourceSelector data from a string data
func GetResourceSelector(stringData string) (*rsML.ResourceSelector, error) {
	data := &rsML.ResourceSelector{}

	labels := GetAsMap(stringData)

	typeQuickID := true
	// verify it is type of quick id or labels
	quickID, ok := labels[rsML.KeyResourceQuickID]
	if !ok {
		if _, found := labels[rsML.KeyResourceLabels]; found {
			typeQuickID = false
		} else {
			return nil, fmt.Errorf("invalid input: [%s]", stringData)
		}
	}
	data.QuickID = quickID

	if !typeQuickID {
		resourceType, ok := labels[rsML.KeyResourceType]
		if !ok {
			return nil, fmt.Errorf("resource type not available: [%s]", stringData)
		}
		data.ResourceType = resourceType
	}

	payload, _ := labels[rsML.KeyResourcePayload]
	preDelay, _ := labels[rsML.KeyResourcePreDelay]
	selector, _ := labels[rsML.KeyResourceSelector]
	data.Payload = payload
	data.PreDelay = preDelay
	data.Selector = selector

	// delete known keys from labels
	delete(labels, rsML.KeyResourceQuickID)
	delete(labels, rsML.KeyResourceLabels)
	delete(labels, rsML.KeyResourceType)
	delete(labels, rsML.KeyResourcePayload)
	delete(labels, rsML.KeyResourcePreDelay)
	delete(labels, rsML.KeyResourceSelector)
	data.Labels = labels

	return data, nil
}
