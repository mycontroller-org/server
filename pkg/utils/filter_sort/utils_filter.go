package helper

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
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
			match = CompareString(toString(value), filter.Operator, filter.Value)

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
			_stringSlice = append(_stringSlice, toString(val))
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
	valueString := toString(value)
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return toString(filterValue) == valueString
	case stgml.OperatorNotEqual:
		return toString(filterValue) != valueString
	case stgml.OperatorRegex:
		expression := fmt.Sprintf("(?i)%s", toString(filterValue))
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
		boolSlice = append(boolSlice, toBool(val))
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
func CompareBool(value interface{}, operator string, filterValue interface{}) bool {
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return toBool(value) == toBool(filterValue)
	case stgml.OperatorNotEqual:
		return toBool(value) != toBool(filterValue)
	case stgml.OperatorExists:
		return len(toString(value)) > 0
	case stgml.OperatorIn, stgml.OperatorNotIn:
		return VerifyBoolSlice(toBool(value), operator, filterValue)
	}
	return false
}

// CompareFloat compares float
func CompareFloat(value interface{}, operator string, filterValue interface{}) bool {
	valueFloat := toFloat(value)
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return valueFloat == toFloat(filterValue)

	case stgml.OperatorNotEqual:
		return valueFloat != toFloat(filterValue)

	case stgml.OperatorGreaterThan:
		return valueFloat > toFloat(filterValue)

	case stgml.OperatorGreaterThanEqual:
		return valueFloat >= toFloat(filterValue)

	case stgml.OperatorLessThan:
		return valueFloat < toFloat(filterValue)

	case stgml.OperatorLessThanEqual:
		return valueFloat <= toFloat(filterValue)

	case stgml.OperatorIn, stgml.OperatorNotIn, stgml.OperatorRangeIn, stgml.OperatorRangeNotIn:
		return VerifyFloatSlice(valueFloat, operator, filterValue)
	}
	return false
}

// VerifyFloatSlice implementation
func VerifyFloatSlice(value float64, operator string, filterValue interface{}) bool {
	floatSlice, ok := filterValue.([]float64)
	if !ok {
		genericSlice, genericOk := filterValue.([]interface{})
		if !genericOk {
			return false
		}
		_floatSlice := make([]float64, 0)
		for _, val := range genericSlice {
			_floatSlice = append(_floatSlice, toFloat(val))
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

func toString(data interface{}) string {
	value, ok := data.(string)
	if !ok {
		value = fmt.Sprintf("%v", data)
	}
	return value
}

func toBool(data interface{}) bool {
	value, ok := data.(bool)
	if !ok {
		value = strings.ToLower(fmt.Sprintf("%v", data)) == "true"
	}
	return value
}

func toFloat(data interface{}) float64 {
	value, ok := data.(float64)
	if !ok {
		strValue := fmt.Sprintf("%v", data)
		parsedValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return 0
		}
		return parsedValue
	}
	return value
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
