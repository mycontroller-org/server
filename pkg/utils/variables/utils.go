package variables

import (
	"encoding/base64"
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// UpdateParameters updates parameter templates
func UpdateParameters(logger *zap.Logger, variables map[string]interface{}, parameters map[string]string, templateEngine types.TemplateEngine) map[string]string {
	updatedParameters := make(map[string]string)
	for name, value := range parameters {
		// load suplied string, this will be passed, if there is an error
		updatedParameters[name] = value

		genericData := handlerTY.GenericData{}
		err := json.Unmarshal([]byte(value), &genericData)
		if err == nil {
			// unpack base64 to normal string
			yamlBytes, err := base64.StdEncoding.DecodeString(genericData.Data)
			if err != nil {
				logger.Error("error on converting parameter data", zap.String("name", name), zap.Error(err))
				continue
			}
			// execute template
			updatedValue, err := templateEngine.Execute(string(yamlBytes), variables)
			if err != nil {
				logger.Error("error on executing template", zap.Error(err), zap.String("name", name), zap.Any("value", string(yamlBytes)))
				updatedParameters[name] = err.Error()
				continue
			}

			// update the disabled value via template
			updatedDisable, err := templateEngine.Execute(genericData.Disabled, variables)
			if err != nil {
				logger.Error("error on executing template, to update disabled value", zap.Error(err), zap.String("name", name), zap.Any("value", genericData.Disabled))
				updatedParameters[name] = err.Error()
				continue
			}
			genericData.Disabled = updatedDisable

			// repack string to base64 string
			genericData.Data = base64.StdEncoding.EncodeToString([]byte(updatedValue))

			// if it is a webhook data and customData not enabled, update variables on the data field
			if genericData.Type == handlerTY.DataTypeWebhook {
				webhookData := handlerTY.WebhookData{}
				err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &webhookData)
				if err != nil {
					logger.Error("error on converting webhook data", zap.Error(err), zap.String("name", name))
					continue
				}
				if webhookData.Method != http.MethodGet && !webhookData.CustomData {
					webhookData.Data = variables
					updatedString, err := yamlUtils.MarshalBase64Yaml(webhookData)
					if err != nil {
						logger.Error("error on converting webhook data to yaml", zap.Error(err), zap.String("name", name))
						continue
					}
					genericData.Data = updatedString
				}
			}

			jsonBytes, err := json.Marshal(genericData)
			if err != nil {
				logger.Error("error on converting to json", zap.Error(err), zap.String("name", name))
			}
			updatedParameters[name] = string(jsonBytes)
		} else { // update as a normal text
			updatedValue, err := templateEngine.Execute(value, variables)
			if err != nil {
				logger.Warn("error on executing template", zap.Error(err), zap.String("name", name), zap.Any("value", value))
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
