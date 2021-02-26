package templateutils

import (
	"bytes"
	"html/template"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	"github.com/mycontroller-org/backend/v2/pkg/version"
	"go.uber.org/zap"
)

var funcMap = template.FuncMap{
	"now":        time.Now,
	"version":    version.Get,
	"marshal":    marshal,
	"toJson":     marshal,
	"bySelector": bySelector,
	"ternary":    ternary,
}

func ternary(data interface{}, trueValue, falseValue string) string {
	if utils.ToBool(data) {
		return trueValue
	}
	return falseValue
}

func bySelector(data interface{}, selector string) template.JS {
	result := ""
	_, value, err := helper.GetValueByKeyPath(data, selector)
	if err != nil {
		zap.L().Warn("error on getting a value", zap.Error(err), zap.String("selector", selector), zap.Any("data", data))
		result = err.Error()
	} else {
		stringResult, ok := value.(string)
		if ok {
			result = stringResult
		} else {
			stringResult, err := json.Marshal(value)
			if err != nil {
				zap.L().Warn("error on converting to json", zap.Error(err), zap.String("selector", selector), zap.Any("data", data))
			} else {
				result = string(stringResult)
			}
		}
	}
	return template.JS(result)
}

func marshal(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}

// Execute a template
func Execute(templateText string, data interface{}) (string, error) {
	tpl, err := template.New("base").Funcs(funcMap).Parse(templateText)
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
