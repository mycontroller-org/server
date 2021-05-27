package variables

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	dataRepositoryAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/backend/v2/pkg/api/schedule"
	sourceAPI "github.com/mycontroller-org/backend/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	"github.com/mycontroller-org/backend/v2/pkg/json"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"gopkg.in/yaml.v2"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/utils"

	cloneUtil "github.com/mycontroller-org/backend/v2/pkg/utils/clone"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	quickIdUL "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	templateUtils "github.com/mycontroller-org/backend/v2/pkg/utils/template"
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

	// clone variables
	clonedVariables := cloneUtil.Clone(variables)

	backToVariables, ok := clonedVariables.(map[string]interface{})
	if ok {
		// descrypt the secrets, tokens
		err := cloneUtil.UpdateSecrets(backToVariables, false)
		if err != nil {
			return nil, err
		}

		return backToVariables, nil
	}

	zap.L().Error("error on clone, returning variables as is")
	return variables, nil
}

func getEntity(name, stringValue string) interface{} {
	genericData := handlerML.GenericData{}
	err := json.Unmarshal([]byte(stringValue), &genericData)
	if err != nil {
		// if error happens, this could be a normal string or templated string
		// try it as template
		formattedString, err := templateUtils.Execute(stringValue, nil)
		if err != nil {
			return fmt.Sprintf("error on executing template. name:%s, template:%s, error:%s", name, stringValue, err.Error())
		}
		return formattedString
	}

	// process only for resource type data
	if !strings.HasPrefix(genericData.Type, handlerML.DataTypeResource) {
		return stringValue
	}

	rsData := handlerML.ResourceData{}
	err = UnmarshalBase64Yaml(genericData.Data, &rsData)
	if err != nil {
		zap.L().Error("error on loading resource data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
		return err.Error()
	}

	if rsData.QuickID != "" {
		return getByQuickID(name, &rsData)
	} else if rsData.ResourceType != "" && len(rsData.Labels) > 0 {
		return getByLabels(name, &rsData)
	}

	return stringValue
}

func getByLabels(name string, rsData *handlerML.ResourceData) interface{} {

	apiImpl := genericAPI{}
	switch {
	case utils.ContainsString(quickIdUL.QuickIDGateway, rsData.ResourceType):
		apiImpl.List = gatewayAPI.List

	case utils.ContainsString(quickIdUL.QuickIDNode, rsData.ResourceType):
		apiImpl.List = nodeAPI.List

	case utils.ContainsString(quickIdUL.QuickIDSource, rsData.ResourceType):
		apiImpl.List = sourceAPI.List

	case utils.ContainsString(quickIdUL.QuickIDField, rsData.ResourceType):
		apiImpl.List = fieldAPI.List

	case utils.ContainsString(quickIdUL.QuickIDTask, rsData.ResourceType):
		apiImpl.List = taskAPI.List

	case utils.ContainsString(quickIdUL.QuickIDSchedule, rsData.ResourceType):
		apiImpl.List = scheduleAPI.List

	case utils.ContainsString(quickIdUL.QuickIDHandler, rsData.ResourceType):
		apiImpl.List = handlerAPI.List

	case utils.ContainsString(quickIdUL.QuickIDDataRepository, rsData.ResourceType):
		apiImpl.List = dataRepositoryAPI.List
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
			zap.L().Warn("error on getting label based entity", zap.Error(err), zap.String("name", name), zap.Any("rsData", rsData))
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

func getByQuickID(name string, rsData *handlerML.ResourceData) interface{} {
	quickID := fmt.Sprintf("%s:%s", rsData.ResourceType, rsData.QuickID)
	resourceType, keys, err := quickIdUL.EntityKeyValueMap(quickID)
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

	case utils.ContainsString(quickIdUL.QuickIDSource, resourceType):
		item, err := sourceAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySourceID])
		entity = item
		if err != nil {
			zap.L().Warn("source not available", zap.Any("keys", keys))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDField, resourceType):
		item, err := fieldAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySourceID], keys[model.KeyFieldID])
		entity = item
		if err != nil {
			zap.L().Warn("field not available", zap.Any("keys", keys), zap.Error(err))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDTask, resourceType):
		item, err := taskAPI.GetByID(keys[model.KeyID])
		entity = item
		if err != nil {
			zap.L().Warn("task not available", zap.Any("keys", keys))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDSchedule, resourceType):
		item, err := scheduleAPI.GetByID(keys[model.KeyID])
		entity = item
		if err != nil {
			zap.L().Warn("schedule not available", zap.Any("keys", keys))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDHandler, resourceType):
		item, err := handlerAPI.GetByID(keys[model.KeyID])
		entity = item
		if err != nil {
			zap.L().Warn("handler not available", zap.Any("keys", keys))
			return nil
		}

	case utils.ContainsString(quickIdUL.QuickIDDataRepository, resourceType):
		item, err := dataRepositoryAPI.GetByID(keys[model.KeyID])
		entity = item
		if err != nil {
			zap.L().Warn("data not available in data repository", zap.Any("keys", keys))
			return nil
		}

	default:
		data, err := templateUtils.Execute(keys[model.KeyTemplate], nil)
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
		_, value, err := helper.GetValueByKeyPath(entity, rsData.Selector)
		if err != nil {
			zap.L().Error("error on getting data from a given selector", zap.Error(err), zap.String("selector", rsData.Selector), zap.Any("entity", entity))
			return nil
		}
		return value
	}
	return entity
}

// UpdateParameters updates parmaeter templates
func UpdateParameters(variables map[string]interface{}, parameters map[string]string) map[string]string {
	updatedParameters := make(map[string]string)
	for name, value := range parameters {
		// load suplied string, this will be passed, if there is an error
		updatedParameters[name] = value

		genericData := handlerML.GenericData{}
		err := json.Unmarshal([]byte(value), &genericData)
		if err == nil {
			// unpack base64 to normal string
			yamlBytes, err := base64.StdEncoding.DecodeString(genericData.Data)
			if err != nil {
				zap.L().Error("error on converting parameter data", zap.String("name", name), zap.Error(err))
				continue
			}
			// execute template
			updatedValue, err := templateUtils.Execute(string(yamlBytes), variables)
			if err != nil {
				zap.L().Error("error on executing template", zap.Error(err), zap.String("name", name), zap.Any("value", string(yamlBytes)))
				updatedParameters[name] = err.Error()
				continue
			}

			// update the disabled value via template
			updatedDisable, err := templateUtils.Execute(genericData.Disabled, variables)
			if err != nil {
				zap.L().Error("error on executing template, to update disabled value", zap.Error(err), zap.String("name", name), zap.Any("value", genericData.Disabled))
				updatedParameters[name] = err.Error()
				continue
			}
			genericData.Disabled = updatedDisable

			// repack string to base64 string
			genericData.Data = base64.StdEncoding.EncodeToString([]byte(updatedValue))

			// if it is a webhook data and customData not enabled, update variables on the data field
			if genericData.Type == handlerML.DataTypeWebhook {
				webhookData := handlerML.WebhookData{}
				err = UnmarshalBase64Yaml(genericData.Data, &webhookData)
				if err != nil {
					zap.L().Error("error on converting webhook data", zap.Error(err), zap.String("name", name))
					continue
				}
				if webhookData.Method != http.MethodGet && !webhookData.CustomData {
					webhookData.Data = variables
					updatedString, err := MarshalBase64Yaml(webhookData)
					if err != nil {
						zap.L().Error("error on converting webhook data to yaml", zap.Error(err), zap.String("name", name))
						continue
					}
					genericData.Data = updatedString
				}
			}

			jsonBytes, err := json.Marshal(genericData)
			if err != nil {
				zap.L().Error("error on converting to json", zap.Error(err), zap.String("name", name))
			}
			updatedParameters[name] = string(jsonBytes)
		} else { // update as a normal text
			updatedValue, err := templateUtils.Execute(value, variables)
			if err != nil {
				zap.L().Warn("error on executing template", zap.Error(err), zap.String("name", name), zap.Any("value", value))
				updatedParameters[name] = err.Error()
				continue
			}
			updatedParameters[name] = updatedValue
		}

	}
	return updatedParameters
}

// Merge variables and extra variables
func Merge(variables map[string]interface{}, extra map[string]interface{}) map[string]interface{} {
	finalMap := make(map[string]interface{})

	if len(variables) > 0 {
		for name, value := range variables { // update variables
			finalMap[name] = value
		}
	}

	if len(extra) > 0 {
		for name, value := range extra { // update extra
			finalMap[name] = value
		}
	}

	return finalMap
}

// UnmarshalBase64Yaml converts base64 data into given interface
func UnmarshalBase64Yaml(base64String string, out interface{}) error {
	yamlBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(yamlBytes, out)
}

// MarshalBase64Yaml converts interface to base64
func MarshalBase64Yaml(in interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(in)
	if err != nil {
		return "", err
	}
	base64string := base64.StdEncoding.EncodeToString(yamlBytes)
	return base64string, nil
}
