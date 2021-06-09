package templateutils

import (
	"bytes"
	"html/template"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/api/sunrise"
	"github.com/mycontroller-org/backend/v2/pkg/json"
	converterUtils "github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	"github.com/mycontroller-org/backend/v2/pkg/version"
	"go.uber.org/zap"
)

var funcMap = template.FuncMap{
	"now":     time.Now,
	"version": version.Get,
	"marshal": marshal,
	"toJson":  marshal,
	"keyPath": byKeyPath,
	"ternary": ternary,
	"sunrise": sunrise.GetSunriseTime,
	"sunset":  sunrise.GetSunsetTime,
}

func ternary(data interface{}, trueValue, falseValue string) string {
	if converterUtils.ToBool(data) {
		return trueValue
	}
	return falseValue
}

func byKeyPath(data interface{}, keyPath string) template.JS {
	result := ""
	_, value, err := helper.GetValueByKeyPath(data, keyPath)
	if err != nil {
		zap.L().Warn("error on getting a value", zap.Error(err), zap.String("keyPath", keyPath), zap.Any("data", data))
		result = err.Error()
	} else {
		stringResult, ok := value.(string)
		if ok {
			result = stringResult
		} else {
			stringResult, err := json.Marshal(value)
			if err != nil {
				zap.L().Warn("error on converting to json", zap.Error(err), zap.String("keyPath", keyPath), zap.Any("data", data))
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
