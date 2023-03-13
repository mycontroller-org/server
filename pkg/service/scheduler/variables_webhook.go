package scheduler

import (
	"net/http"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	"go.uber.org/zap"
)

const timeout = time.Second * 10

func (svc *SchedulerService) loadWebhookVariables(scheduleID string, config schedulerTY.CustomVariableConfig, variables map[string]interface{}) map[string]interface{} {
	whCfg := config.Webhook
	client := httpclient.GetClient(whCfg.Insecure, timeout)
	if !whCfg.IncludeConfig {
		delete(variables, types.KeySchedule)
	}

	if whCfg.Method == "" {
		whCfg.Method = http.MethodPost
	}

	res, err := client.ExecuteJson(whCfg.URL, whCfg.Method, whCfg.Headers, whCfg.QueryParameters, variables, 0)
	responseStatusCode := 0
	if res != nil {
		responseStatusCode = res.StatusCode
	}
	if err != nil {
		svc.logger.Error("error on executing webhook", zap.Error(err), zap.String("scheduleID", scheduleID), zap.String("url", whCfg.URL), zap.Int("responseStatusCode", responseStatusCode))
		return nil
	}

	resultMap := make(map[string]interface{})

	err = json.Unmarshal(res.Body, &resultMap)
	if err != nil {
		svc.logger.Error("error on converting to json", zap.Error(err), zap.String("response", res.StringBody()))
		return nil
	}

	return resultMap
}
