package variables

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	dataRepositoryAPI "github.com/mycontroller-org/server/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	handlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	sourceAPI "github.com/mycontroller-org/server/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	"github.com/mycontroller-org/server/v2/pkg/json"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/utils"

	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"

	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	quickIdUL "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	templateUtils "github.com/mycontroller-org/server/v2/pkg/utils/template"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

const (
	webhookTimeout = time.Second * 10
)

type genericAPI struct {
	List func(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error)
}

// LoadVariables loads all the defined variables
func LoadVariables(variablesPreMap map[string]string, secret string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	for name, stringValue := range variablesPreMap {
		value := getEntity(name, stringValue)
		if value == nil {
			return nil, fmt.Errorf("failed to load a variable. name: %s, keyPath:%s", name, stringValue)
		}
		variables[name] = value
	}

	// clone variables
	clonedVariables := cloneUtil.Clone(variables)

	backToVariables, ok := clonedVariables.(map[string]interface{})
	if ok {
		// descrypt the secrets, tokens
		err := cloneUtil.UpdateSecrets(backToVariables, secret, "", false, cloneUtil.DefaultSpecialKeys)
		if err != nil {
			return nil, err
		}

		return backToVariables, nil
	}

	zap.L().Error("error on clone, returning variables as is")
	return variables, nil
}

func getEntity(name, stringValue string) interface{} {
	genericData := handlerType.GenericData{}
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
	if !strings.HasPrefix(genericData.Type, handlerType.DataTypeResource) &&
		genericData.Type != handlerType.DataTypeWebhook {
		return stringValue
	}

	// calls webhook and loads the response as is
	if genericData.Type == handlerType.DataTypeWebhook {
		webhookCfg := handlerType.WebhookData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &webhookCfg)
		if err != nil {
			zap.L().Error("error on loading webhook data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
			return err.Error()
		}
		return getWebhookData(name, &webhookCfg)

	} else {
		rsData := handlerType.ResourceData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &rsData)
		if err != nil {
			zap.L().Error("error on loading resource data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
			return err.Error()
		}

		if rsData.QuickID != "" {
			return getByQuickID(name, &rsData)
		} else if rsData.ResourceType != "" && len(rsData.Labels) > 0 {
			return getByLabels(name, &rsData)
		}
	}

	return stringValue
}

func getWebhookData(name string, whCfg *handlerType.WebhookData) interface{} {
	client := httpclient.GetClient(whCfg.InsecureSkipVerify, webhookTimeout)

	if whCfg.Method == "" {
		whCfg.Method = http.MethodGet
	}

	res, resBody, err := client.Request(whCfg.Server, whCfg.Method, whCfg.Headers, whCfg.QueryParameters, whCfg.Data, whCfg.ResponseCode)
	responseStatusCode := 0
	if res != nil {
		responseStatusCode = res.StatusCode
	}
	if err != nil {
		zap.L().Error("error on executing webhook", zap.Error(err), zap.String("variableName", name), zap.String("server", whCfg.Server), zap.Int("responseStatusCode", responseStatusCode))
		return nil
	}

	resultMap := make(map[string]interface{})

	err = json.Unmarshal(resBody, &resultMap)
	if err != nil {
		zap.L().Error("error on converting to json", zap.Error(err), zap.String("response", string(resBody)))
		return nil
	}
	return resultMap
}

func getByLabels(name string, rsData *handlerType.ResourceData) interface{} {

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
		filters := make([]stgType.Filter, 0)
		for key, value := range rsData.Labels {
			filter := stgType.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: stgType.OperatorEqual, Value: value}
			filters = append(filters, filter)
		}
		pagination := &stgType.Pagination{Limit: 1} // limit to one element
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
			if rsData.KeyPath != "" {
				_, value, err := helper.GetValueByKeyPath(entity, rsData.KeyPath)
				if err != nil {
					zap.L().Warn("error on getting data from a given keyPath", zap.Error(err), zap.String("keyPath", rsData.KeyPath), zap.Any("entity", entity))
					return nil
				}
				return value
			}
			return entity
		}
	}

	return nil
}

func getByQuickID(name string, rsData *handlerType.ResourceData) interface{} {
	quickID := fmt.Sprintf("%s:%s", rsData.ResourceType, rsData.QuickID)
	resourceType, keys, err := quickIdUL.EntityKeyValueMap(quickID)
	if err != nil {
		zap.L().Warn("failed to parse variable", zap.Any("name", name), zap.Any("data", rsData), zap.Error(err))
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

	if rsData.KeyPath != "" {
		_, value, err := helper.GetValueByKeyPath(entity, rsData.KeyPath)
		if err != nil {
			zap.L().Error("error on getting data from a given keyPath", zap.Error(err), zap.String("keyPath", rsData.KeyPath), zap.Any("entity", entity))
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

		genericData := handlerType.GenericData{}
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
			if genericData.Type == handlerType.DataTypeWebhook {
				webhookData := handlerType.WebhookData{}
				err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &webhookData)
				if err != nil {
					zap.L().Error("error on converting webhook data", zap.Error(err), zap.String("name", name))
					continue
				}
				if webhookData.Method != http.MethodGet && !webhookData.CustomData {
					webhookData.Data = variables
					updatedString, err := yamlUtils.MarshalBase64Yaml(webhookData)
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
