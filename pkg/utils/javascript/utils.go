package javascript

import (
	"github.com/dop251/goja"
	"go.uber.org/zap"
)

// Execute a given javascript
func Execute(scriptString string, variables map[string]interface{}) (interface{}, error) {
	rt := goja.New()
	// enable this line if we want to use supplied object as json
	// GoLang func call will not be available, if json enabled
	// rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	for name, value := range variables {
		rt.Set(name, value)
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
