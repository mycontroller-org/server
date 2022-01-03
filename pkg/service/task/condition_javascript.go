package task

import (
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

func isTriggeredJavascript(taskID string, config taskTY.EvaluationConfig, variables map[string]interface{}) (map[string]interface{}, bool) {
	result, err := javascript.Execute(config.Javascript, variables)
	if err != nil {
		zap.L().Error("error on executing script", zap.Error(err), zap.String("taskID", taskID), zap.String("script", config.Javascript))
		return nil, false
	}
	// data, _ := json.Marshal(result)
	// zap.L().Info("variables", zap.Any("vars", string(data)))
	if resultMap, ok := result.(map[string]interface{}); ok {
		isTriggered, isTriggeredFound := resultMap[taskTY.KeyIsTriggered]
		if isTriggeredFound {
			return resultMap, converterUtils.ToBool(isTriggered)
		}
		return resultMap, false
	}
	return nil, converterUtils.ToBool(result)
}
