package task

import (
	"net/http"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	converterUtils "github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
	httpclient "github.com/mycontroller-org/backend/v2/pkg/utils/http_client_json"
	"go.uber.org/zap"
)

func isTriggeredWebhook(taskID string, config taskML.EvaluationConfig, variables map[string]interface{}) (map[string]interface{}, bool) {
	whCfg := config.Webhook
	client := httpclient.GetClient(whCfg.InsecureSkipVerify)
	if !whCfg.IncludeConfig {
		delete(variables, model.KeyTask)
	}

	if whCfg.Method == "" {
		whCfg.Method = http.MethodPost
	}

	res, resBody, err := client.Request(whCfg.URL, whCfg.Method, whCfg.Headers, whCfg.QueryParameters, variables, 0)
	responseStatusCode := 0
	if res != nil {
		responseStatusCode = res.StatusCode
	}
	if err != nil {
		zap.L().Error("error on executing webhook", zap.Error(err), zap.String("taskID", taskID), zap.String("url", whCfg.URL), zap.Int("responseStatusCode", responseStatusCode))
		return nil, false
	}

	resultMap := make(map[string]interface{})

	err = json.Unmarshal(resBody, &resultMap)
	if err != nil {
		zap.L().Error("error on converting to json", zap.Error(err), zap.String("response", string(resBody)))
		return nil, converterUtils.ToBool(string(resBody))
	}

	zap.L().Debug("webhook response", zap.String("taskID", taskID), zap.Any("response", resultMap))
	if len(resultMap) > 0 {
		isTriggered, isTriggeredFound := resultMap[taskML.KeyIsTriggered]
		if isTriggeredFound {
			return resultMap, converterUtils.ToBool(isTriggered)
		}
		return resultMap, false
	}

	return nil, false
}
