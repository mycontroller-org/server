package javascript

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"

	"go.uber.org/zap"
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
			zap.L().Warn("error on setting a value", zap.String("name", name), zap.Any("value", value))
		}
	}
	zap.L().Debug("executing script", zap.Any("variables", variables), zap.String("scriptString", scriptString))
	response, err := rt.RunString(string(scriptString))
	if err != nil {
		return nil, err
	}

	output := response.Export()
	zap.L().Debug("executed script", zap.Any("variables", variables), zap.String("scriptString", scriptString), zap.Any("output", output))

	return output, nil
}
