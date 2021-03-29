package helper

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	converterUtils "github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// FilterByStringMap filter
func FilterByStringMap(entities []interface{}, filtersMap map[string]string) []interface{} {
	filters := make([]stgml.Filter, 0)
	for k, v := range filtersMap {
		filters = append(filters, stgml.Filter{Key: k, Operator: stgml.OperatorEqual, Value: v})
	}
	if len(filters) == 0 {
		return entities
	}
	return Filter(entities, filters, false)
}

// Filter filters the given slice
func Filter(entities []interface{}, filters []stgml.Filter, returnSingle bool) []interface{} {
	filteredEntities := make([]interface{}, 0)
	if len(filters) == 0 {
		return entities
	}

	for _, entity := range entities {
		match := IsMatching(entity, filters)
		if match {
			filteredEntities = append(filteredEntities, entity)
		}
		if returnSingle && len(filteredEntities) > 0 {
			return filteredEntities[:1]
		}
	}
	return filteredEntities
}

// IsMatching returns matching status
func IsMatching(entity interface{}, filters []stgml.Filter) bool {
	match := true
	for index := 0; index < len(filters); index++ {
		filter := filters[index]
		valKind, value, err := GetValueByKeyPath(entity, filter.Key)
		if err != nil {
			//zap.L().Debug("failed to get value", zap.Any("filter", filter), zap.Error(err))
			match = false
			break
		}

		switch valKind {
		case reflect.String:
			match = CompareString(converterUtils.ToString(value), filter.Operator, filter.Value)

		case reflect.Bool:
			match = CompareBool(value, filter.Operator, filter.Value)

		default:
			match = false
		}

		if !match {
			break
		}
	}
	return match
}

// VerifyStringSlice implementation
func VerifyStringSlice(value string, operator string, filterValue interface{}) bool {
	stringSlice, ok := filterValue.([]string)
	if !ok {
		genericSlice, genericOk := filterValue.([]interface{})
		if !genericOk {
			return false
		}
		_stringSlice := make([]string, 0)
		for _, val := range genericSlice {
			_stringSlice = append(_stringSlice, converterUtils.ToString(val))
		}
		stringSlice = _stringSlice
	}

	switch operator {
	case stgml.OperatorIn:
		for _, fValue := range stringSlice {
			if value == fValue {
				return true
			}
		}
	case stgml.OperatorNotIn:
		for _, fValue := range stringSlice {
			if value == fValue {
				return false
			}
		}
		return true
	}
	return false
}

// CompareString compares strings
func CompareString(value interface{}, operator string, filterValue interface{}) bool {
	valueString := converterUtils.ToString(value)
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return converterUtils.ToString(filterValue) == valueString
	case stgml.OperatorNotEqual:
		return converterUtils.ToString(filterValue) != valueString
	case stgml.OperatorRegex:
		expression := fmt.Sprintf("(?i)%s", converterUtils.ToString(filterValue))
		compiled, err := regexp.Compile(expression)
		if err != nil {
			return false
		}
		return compiled.MatchString(valueString)
	case stgml.OperatorExists:
		return valueString != ""
	case stgml.OperatorIn, stgml.OperatorNotIn:
		return VerifyStringSlice(valueString, operator, filterValue)
	}
	return false
}

// VerifyBoolSlice implementation
func VerifyBoolSlice(value bool, operator string, filterValue interface{}) bool {
	genericSlice, ok := filterValue.([]interface{})
	if !ok {
		return false
	}
	boolSlice := make([]bool, 0)
	for _, val := range genericSlice {
		boolSlice = append(boolSlice, converterUtils.ToBool(val))
	}

	switch operator {
	case stgml.OperatorIn:
		for _, fValue := range boolSlice {
			if value == fValue {
				return true
			}
		}
	case stgml.OperatorNotIn:
		for _, fValue := range boolSlice {
			if value == fValue {
				return false
			}
		}
		return true
	}
	return false
}

// CompareBool compares strings
func CompareBool(value interface{}, operator string, expectedValue interface{}) bool {
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return converterUtils.ToBool(value) == converterUtils.ToBool(expectedValue)
	case stgml.OperatorNotEqual:
		return converterUtils.ToBool(value) != converterUtils.ToBool(expectedValue)
	case stgml.OperatorExists:
		return len(converterUtils.ToString(value)) > 0
	case stgml.OperatorIn, stgml.OperatorNotIn:
		return VerifyBoolSlice(converterUtils.ToBool(value), operator, expectedValue)
	}
	return false
}

// CompareFloat compares float
func CompareFloat(value interface{}, operator string, expectedValue interface{}) bool {
	valueFloat := converterUtils.ToFloat(value)
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return valueFloat == converterUtils.ToFloat(expectedValue)

	case stgml.OperatorNotEqual:
		return valueFloat != converterUtils.ToFloat(expectedValue)

	case stgml.OperatorGreaterThan:
		return valueFloat > converterUtils.ToFloat(expectedValue)

	case stgml.OperatorGreaterThanEqual:
		return valueFloat >= converterUtils.ToFloat(expectedValue)

	case stgml.OperatorLessThan:
		return valueFloat < converterUtils.ToFloat(expectedValue)

	case stgml.OperatorLessThanEqual:
		return valueFloat <= converterUtils.ToFloat(expectedValue)

	case stgml.OperatorIn, stgml.OperatorNotIn, stgml.OperatorRangeIn, stgml.OperatorRangeNotIn:
		return VerifyFloatSlice(valueFloat, operator, expectedValue)
	}
	return false
}

// VerifyFloatSlice implementation
func VerifyFloatSlice(value float64, operator string, expectedValue interface{}) bool {
	floatSlice, ok := expectedValue.([]float64)
	if !ok {
		genericSlice, genericOk := expectedValue.([]interface{})
		if !genericOk {
			return false
		}
		_floatSlice := make([]float64, 0)
		for _, val := range genericSlice {
			_floatSlice = append(_floatSlice, converterUtils.ToFloat(val))
		}
		floatSlice = _floatSlice
	}

	switch operator {
	case stgml.OperatorIn:
		for _, fValue := range floatSlice {
			if value == fValue {
				return true
			}
		}

	case stgml.OperatorNotIn:
		for _, fValue := range floatSlice {
			if value == fValue {
				return false
			}
		}
		return true

	case stgml.OperatorRangeIn:
		if len(floatSlice) != 2 {
			return false
		}
		lowRange := floatSlice[0]
		highRange := floatSlice[1]
		return value > lowRange && value < highRange

	case stgml.OperatorRangeNotIn:
		if len(floatSlice) != 2 {
			return false
		}
		lowRange := floatSlice[0]
		highRange := floatSlice[1]
		return value < lowRange || value > highRange

	}
	return false
}

// IsMine verifies the supplied id and labels with valid list
func IsMine(validIDs []string, validLabels cmap.CustomStringMap, targetID string, targetLabels cmap.CustomStringMap) bool {
	if len(validIDs) == 0 {
		if len(validLabels) == 0 {
			return true
		}
		for key, value := range validLabels {
			receivedValue, found := targetLabels[key]
			if !found {
				return false
			}
			if value != receivedValue {
				return false
			}
		}
		return true
	}

	for _, id := range validIDs {
		if id == targetID {
			return true
		}
	}
	return false

}
