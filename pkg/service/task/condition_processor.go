package task

import (
	"fmt"
	"reflect"
	"strings"

	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	"go.uber.org/zap"
)

func isTriggered(rule taskML.Rule, variables map[string]interface{}) bool {
	if len(rule.Conditions) == 0 {
		return true
	}

	zap.L().Debug("isTriggered", zap.Any("conditions", rule.Conditions), zap.Any("variables", variables))

	for index := 0; index < len(rule.Conditions); index++ {
		condition := rule.Conditions[index]
		value, err := getValueByVariableName(variables, condition.Variable)
		if err != nil {
			zap.L().Warn("error on getting a variable", zap.Error(err))
			return false
		}
		valid := isMatching(value, condition.Operator, condition.Value)

		if rule.MatchAll && !valid {
			zap.L().Debug("condition failed", zap.Any("c", condition), zap.Any("variables", variables))
			return false
		}

		if !rule.MatchAll && valid {
			zap.L().Debug("condition passed", zap.Any("c", condition), zap.Any("variables", variables))
			return true
		}
	}

	if rule.MatchAll {
		return true
	}
	return false
}

func getValueByVariableName(variables map[string]interface{}, variableName string) (interface{}, error) {
	name := variableName
	selector := ""
	if strings.Contains(variableName, ".") {
		keys := strings.SplitN(variableName, ".", 2)
		name = keys[0]
		selector = keys[1]
	}

	entity, found := variables[name]
	if !found {
		return nil, fmt.Errorf("variable not loaded, variable:%s", name)
	}

	if selector != "" {
		_, value, err := helper.GetValueByKeyPath(entity, selector)
		if err != nil {
			zap.L().Warn("error to get a value for a variable", zap.Error(err), zap.String("variable", name), zap.String("selector", selector))
			return nil, fmt.Errorf("invalid selector. variable:%s, selector:%s", name, selector)
		}
		return value, nil
	}

	return entity, nil
}

func isMatching(value interface{}, operator string, expectedValue interface{}) bool {
	// TODO: fix value type based on supplied operator
	switch reflect.TypeOf(value).Kind() {

	case reflect.String:
		return helper.CompareString(value, operator, expectedValue)

	case reflect.Bool:
		return helper.CompareBool(value, operator, expectedValue)

	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return helper.CompareFloat(value, operator, expectedValue)

	default:
		zap.L().Warn("unsupported type", zap.String("type", reflect.TypeOf(value).String()), zap.Any("value", value))
		return false
	}
}
