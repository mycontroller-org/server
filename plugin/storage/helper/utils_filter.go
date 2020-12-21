package helper

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// Filter filters the given slice
func Filter(entities []interface{}, filters []stgml.Filter, returnSingle bool) []interface{} {
	filteredEntities := make([]interface{}, 0)
	if len(filters) == 0 {
		return entities
	}

	for _, entity := range entities {
		match := true
		for index := 0; index < len(filters); index++ {
			filter := filters[index]

			valKind, value, err := GetValueByKeyPath(filter.Key, entity)
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

		if match {
			filteredEntities = append(filteredEntities, entity)
		}
		if returnSingle && len(filteredEntities) > 0 {
			return filteredEntities[:1]
		}
	}
	return filteredEntities
}

// VerifyStringSlice implementation
func VerifyStringSlice(value string, operator string, filterValue interface{}) bool {
	genericSlice, ok := filterValue.([]interface{})
	if !ok {
		return false
	}
	stringSlice := make([]string, 0)
	for _, val := range genericSlice {
		stringSlice = append(stringSlice, toString(val))
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
func CompareString(value string, operator string, filterValue interface{}) bool {
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return toString(filterValue) == value
	case stgml.OperatorNotEqual:
		return toString(filterValue) != value
	case stgml.OperatorRegex:
		expression := fmt.Sprintf("(?i)%s", toString(filterValue))
		compiled, err := regexp.Compile(expression)
		if err != nil {
			return false
		}
		return compiled.MatchString(value)
	case stgml.OperatorExists:
		return value != ""
	case stgml.OperatorIn, stgml.OperatorNotIn:
		return VerifyStringSlice(value, operator, filterValue)
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
