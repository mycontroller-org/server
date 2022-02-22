package task

import (
	"net/http"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	"go.uber.org/zap"
)

const timeout = time.Second * 10

func isTriggeredWebhook(taskID string, config taskTY.EvaluationConfig, variables map[string]interface{}) (map[string]interface{}, bool) {
	whCfg := config.Webhook
	client := httpclient.GetClient(whCfg.Insecure, timeout)
	if !whCfg.IncludeConfig {
		delete(variables, types.KeyTask)
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
		isTriggered, isTriggeredFound := resultMap[taskTY.KeyIsTriggered]
		if isTriggeredFound {
			return resultMap, converterUtils.ToBool(isTriggered)
		}
		return resultMap, false
	}

	return nil, false
}
