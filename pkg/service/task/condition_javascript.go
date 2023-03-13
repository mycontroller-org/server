package task

import (
	"time"

	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

func (svc *TaskService) isTriggeredJavascript(taskID string, config taskTY.EvaluationConfig, variables map[string]interface{}, scriptTimeout *time.Duration) (map[string]interface{}, bool) {
	// script runs without timeout
	// TODO: include timeout
	result, err := javascript.Execute(svc.logger, config.Javascript, variables, scriptTimeout)
	if err != nil {
		svc.logger.Error("error on executing script", zap.Error(err), zap.String("taskID", taskID), zap.String("script", config.Javascript))
		return nil, false
	}
	// data, _ := json.Marshal(result)
	// svc.logger.Info("variables", zap.Any("vars", string(data)))
	if resultMap, ok := result.(map[string]interface{}); ok {
		isTriggered, isTriggeredFound := resultMap[taskTY.KeyIsTriggered]
		if isTriggeredFound {
			return resultMap, converterUtils.ToBool(isTriggered)
		}
		return resultMap, false
	}
	return nil, converterUtils.ToBool(result)
}
