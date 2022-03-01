package javascript

import (
	"errors"
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"

	"go.uber.org/zap"

	jsHelper "github.com/mycontroller-org/server/v2/pkg/utils/javascript/js_helper"
)

var registry = new(require.Registry) // this can be shared by multiple runtimes

// Execute a given javascript
func Execute(scriptString string, variables map[string]interface{}) (interface{}, error) {
	rt := goja.New()
	// enable this line if we want to use supplied object as json
	// GoLang func call will not be available, if json enabled
	// rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// add console support on javascript
	registry.Enable(rt)
	console.Enable(rt)

	for name, value := range variables {
		err := rt.Set(name, value)
		if err != nil {
			zap.L().Warn("error on setting a value", zap.String("name", name), zap.Any("value", value), zap.Error(err))
		}
	}
	zap.L().Debug("executing script", zap.Any("variables", variables), zap.String("scriptString", scriptString))

	// include helper functions
	err := rt.Set(jsHelper.KeyMcUtils, jsHelper.GetHelperUtils())
	if err != nil {
		zap.L().Warn("error on setting helper functions", zap.Error(err))
	}

	response, err := rt.RunString(string(scriptString))
	if err != nil {
		return nil, err
	}

	output := response.Export()
	zap.L().Debug("executed script", zap.Any("variables", variables), zap.String("scriptString", scriptString), zap.Any("output", output))

	return output, nil
}

// ToMap converts the interface data to map[string]interface{}
func ToMap(data interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, errors.New("empty input")
	}
	mapData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid type:%T", data)
	}
	return mapData, nil
}
