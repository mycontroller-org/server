package variables

import (
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// UpdateParameters updates parameter templates
func UpdateParameters(logger *zap.Logger, variables map[string]interface{}, parameters map[string]interface{}, templateEngine types.TemplateEngine) map[string]interface{} {
	updatedParameters := make(map[string]interface{})
	for name, parameter := range parameters {
		// load supplied string, this will be passed, if there is an error
		updatedParameters[name] = parameter

		// convert it to yaml format
		yamlBytes, err := yaml.Marshal(&parameter)
		if err != nil {
			logger.Error("error on converting parameter data", zap.String("name", name), zap.Error(err))
			continue
		}
		// execute template
		updatedValue, err := templateEngine.Execute(string(yamlBytes), variables)
		if err != nil {
			logger.Error("error on executing template", zap.String("name", name), zap.Any("value", string(yamlBytes)), zap.Error(err))
			continue
		}
		// convert it back to parameter
		updatedParameter := make(map[string]interface{})
		err = yaml.Unmarshal([]byte(updatedValue), &updatedParameter)
		if err != nil {
			logger.Error("error on converting yaml to struct", zap.String("name", name), zap.Any("value", string(yamlBytes)), zap.Error(err))
			continue
		}

		updatedParameterMap := cmap.CustomMap(updatedParameter)
		if updatedParameterMap.GetString(types.KeyType) == handlerTY.DataTypeWebhook {
			webhookData := handlerTY.WebhookData{}
			err := utils.MapToStruct(utils.TagNameNone, updatedParameter, &webhookData)
			if err != nil {
				logger.Error("error on converting into webhook data", zap.Error(err), zap.String("name", name), zap.Any("input", updatedParameter))
				continue
			}

			if webhookData.Method != http.MethodGet && !webhookData.CustomData {
				webhookData.Data = variables
				yamlVariables, err := yaml.Marshal(variables)
				if err != nil {
					logger.Error("error on converting webhook data to yaml", zap.Error(err), zap.String("name", name))
					continue
				}
				webhookData.Data = string(yamlVariables)
				updatedParameter = utils.StructToMap(&webhookData)
			}
		}
		updatedParameters[name] = updatedParameter
	}

	return updatedParameters
}

// Merge variables and extra variables
func Merge(variables, extra map[string]interface{}) map[string]interface{} {
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

func GetWebhookData(logger *zap.Logger, name string, whCfg *handlerTY.WebhookData) interface{} {
	client := httpclient.GetClient(whCfg.Insecure, webhookTimeout)

	if whCfg.Method == "" {
		whCfg.Method = http.MethodGet
	}

	res, err := client.ExecuteJson(whCfg.Server, whCfg.Method, whCfg.Headers, whCfg.QueryParameters, whCfg.Data, whCfg.ResponseCode)
	responseStatusCode := 0
	if res != nil {
		responseStatusCode = res.StatusCode
	}
	if err != nil {
		logger.Error("error on executing webhook", zap.Error(err), zap.String("variableName", name), zap.String("server", whCfg.Server), zap.Int("responseStatusCode", responseStatusCode))
		return nil
	}

	resultMap := make(map[string]interface{})

	err = json.Unmarshal(res.Body, &resultMap)
	if err != nil {
		logger.Error("error on converting to json", zap.Error(err), zap.String("response", res.StringBody()))
		return nil
	}
	return resultMap
}
