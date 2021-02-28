package task

import (
	"encoding/base64"

	"github.com/dop251/goja"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

func isTriggeredJavascript(taskID string, config taskML.EvaluationConfig, variables map[string]interface{}) (map[string]interface{}, bool) {
	base64String := config.JavaScript
	stringScript, err := base64.StdEncoding.DecodeString(base64String)

	rt := goja.New()
	rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	for name, value := range variables {
		rt.Set(name, value)
	}
	response, err := rt.RunString(string(stringScript))
	if err != nil {
		zap.L().Error("error on executing a script", zap.String("taskID", taskID), zap.String("script", string(stringScript)), zap.Error(err))
		return nil, false
	}

	result := response.Export()
	if resultMap, ok := result.(map[string]interface{}); ok {
		isTriggered, found := resultMap[taskML.KeyScriptIsTriggered]
		isTriggeredBool := utils.ToBool(isTriggered)
		if found || isTriggeredBool {
			return resultMap, true
		}
		return resultMap, false
	}
	return nil, utils.ToBool(result)
}
