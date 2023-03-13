package task

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

func (svc *TaskService) isTriggered(rule taskTY.Rule, variables map[string]interface{}) bool {
	if len(rule.Conditions) == 0 {
		return true
	}

	svc.logger.Debug("isTriggered", zap.Any("conditions", rule.Conditions), zap.Any("variables", variables))

	for index := 0; index < len(rule.Conditions); index++ {
		condition := rule.Conditions[index]
		value, err := svc.getValueByVariableName(variables, condition.Variable)
		if err != nil {
			svc.logger.Warn("error on getting a variable", zap.Error(err))
			return false
		}

		expectedValue := condition.Value
		stringValue := converterUtils.ToString(expectedValue)

		// process value as template
		updatedValue, err := svc.variablesEngine.TemplateEngine().Execute(stringValue, variables)
		if err != nil {
			svc.logger.Warn("error on parsing template", zap.Error(err), zap.String("template", stringValue), zap.Any("variables", variables))
		} else {
			expectedValue = updatedValue
		}

		valid := svc.isMatching(value, condition.Operator, expectedValue)

		if rule.MatchAll && !valid {
			svc.logger.Debug("condition failed", zap.Any("condition", condition), zap.Any("variables", variables), zap.Any("expectedValue", expectedValue))
			return false
		}

		if !rule.MatchAll && valid {
			svc.logger.Debug("condition passed", zap.Any("condition", condition), zap.Any("variables", variables), zap.Any("expectedValue", expectedValue))
			return true
		}
	}

	return rule.MatchAll
}

func (svc *TaskService) getValueByVariableName(variables map[string]interface{}, variableName string) (interface{}, error) {
	name := variableName
	keyPath := ""
	if strings.Contains(variableName, ".") {
		keys := strings.SplitN(variableName, ".", 2)
		name = keys[0]
		keyPath = keys[1]
	}

	entity, found := variables[name]
	if !found {
		return nil, fmt.Errorf("variable not loaded, variable:%s", name)
	}

	if keyPath != "" {
		_, value, err := filterUtils.GetValueByKeyPath(entity, keyPath)
		if err != nil {
			svc.logger.Warn("error to get a value for a variable", zap.Error(err), zap.String("variable", name), zap.String("keyPath", keyPath))
			return nil, fmt.Errorf("invalid keyPath. variable:%s, keyPath:%s", name, keyPath)
		}
		return value, nil
	}

	return entity, nil
}

func (svc *TaskService) isMatching(value interface{}, operator string, expectedValue interface{}) bool {
	if operator == "" {
		operator = storageTY.OperatorEqual
	}

	// format value to actual type
	value = svc.formatValue(value)

	var expectedValueUpdated interface{}

	switch operator {

	// convert json string to object, if required
	case storageTY.OperatorIn, storageTY.OperatorNotIn, storageTY.OperatorRangeIn, storageTY.OperatorRangeNotIn:
		stringValue := converterUtils.ToString(expectedValue)
		updated := make([]interface{}, 0)
		err := json.Unmarshal([]byte(stringValue), &updated)
		if err != nil {
			svc.logger.Error("error on converting expectedValue to array format", zap.Error(err), zap.Any("expectedValue", expectedValue))
			return false
		}
		expectedValueUpdated = updated

	default:
		expectedValueUpdated = expectedValue
	}

	svc.logger.Debug("ismatching", zap.String("valueType", reflect.TypeOf(value).Kind().String()), zap.Any("value", value), zap.String("operator", operator), zap.Any("expectedValue", expectedValueUpdated))

	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		return filterUtils.CompareString(value, operator, expectedValueUpdated)

	case reflect.Bool:
		return filterUtils.CompareBool(value, operator, expectedValueUpdated)

	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return filterUtils.CompareFloat(value, operator, expectedValueUpdated)

	default:
		svc.logger.Warn("unsupported type", zap.String("type", reflect.TypeOf(value).String()), zap.Any("value", value))
		return false
	}
}

// tries to convert the string value to float, bool, string
func (svc *TaskService) formatValue(value interface{}) interface{} {
	if reflect.TypeOf(value).Kind() == reflect.String {
		stringValue := converterUtils.ToString(value)
		// can be a float value
		floatValue, err := strconv.ParseFloat(stringValue, 64)
		if err == nil {
			return floatValue
		}

		// can be a bool value
		boolValue, err := strconv.ParseBool(stringValue)
		if err == nil {
			return boolValue
		}

		// return as string
		return value
	}
	return value
}
