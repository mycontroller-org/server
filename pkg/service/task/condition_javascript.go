package task

import (
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

func isTriggeredJavascript(taskID string, config taskML.EvaluationConfig, variables map[string]interface{}) (map[string]interface{}, bool) {
	result, err := javascript.Execute(config.Javascript, variables)
	if err != nil {
		zap.L().Error("error on executing script", zap.Error(err), zap.String("taskID", taskID), zap.String("script", config.Javascript))
		return nil, false
	}
	if resultMap, ok := result.(map[string]interface{}); ok {
		isTriggered, isTriggeredFound := resultMap[taskML.KeyScriptIsTriggered]
		if isTriggeredFound {
			return resultMap, utils.ToBool(isTriggered)
		}
		return resultMap, false
	}
	return nil, utils.ToBool(result)
}
