package templateutils

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	"go.uber.org/zap"
)

type Template struct {
	logger       *zap.Logger
	functionsMap template.FuncMap
}

func New(ctx context.Context, additionalFuncs map[string]interface{}) (*Template, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	t := &Template{
		logger: logger,
	}
	t.functionsMap = t.getFuncMap(additionalFuncs)

	return t, nil
}

// Execute a template
func (t *Template) Execute(templateText string, data interface{}) (string, error) {
	tpl, err := template.New("base").Funcs(t.functionsMap).Parse(templateText)
	if err != nil {
		return "", err
	}
	var tplBuffer bytes.Buffer
	err = tpl.Execute(&tplBuffer, data)
	if err != nil {
		return "", err
	}
	return tplBuffer.String(), nil
}

func (t *Template) getFuncMap(additionalFuncs map[string]interface{}) template.FuncMap {
	// version, sunrise and sunset support removed
	funcMap := template.FuncMap{
		"now":     time.Now,
		"marshal": t.marshal,
		"toJson":  t.marshal,
		"keyPath": t.byKeyPath,
		"ternary": t.ternary,
	}
	for key, value := range additionalFuncs {
		funcMap[key] = value
	}
	return funcMap
}

func (t *Template) ternary(data interface{}, trueValue, falseValue string) string {
	if converterUtils.ToBool(data) {
		return trueValue
	}
	return falseValue
}

func (t *Template) byKeyPath(data interface{}, keyPath string) template.JS {
	result := ""
	_, value, err := filterUtils.GetValueByKeyPath(data, keyPath)
	if err != nil {
		t.logger.Warn("error on getting a value", zap.Error(err), zap.String("keyPath", keyPath), zap.Any("data", data))
		result = err.Error()
	} else {
		stringResult, ok := value.(string)
		if ok {
			result = stringResult
		} else {
			stringResult, err := json.Marshal(value)
			if err != nil {
				t.logger.Warn("error on converting to json", zap.Error(err), zap.String("keyPath", keyPath), zap.Any("data", data))
			} else {
				result = string(stringResult)
			}
		}
	}
	return template.JS(result)
}

func (t *Template) marshal(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}
