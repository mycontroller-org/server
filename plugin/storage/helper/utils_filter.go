package helper

import (
	"fmt"
	"reflect"
	"regexp"

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
				filterValue := fmt.Sprintf("%v", filter.Value)
				match = CompareString(filterValue, filter.Operator, value)

			case reflect.Bool:
				if filterValue, ok := filter.Value.(bool); ok {
					match = CompareBool(filterValue, filter.Operator, value)
				} else {
					match = false
				}
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
func VerifyStringSlice(data interface{}, expected bool, verifyFn func(value string) bool) bool {
	stringSlice, ok := data.([]string)
	if !ok {
		return false
	}
	for _, value := range stringSlice {
		if verifyFn(value) == expected {
			return true
		}
	}
	return false
}

// CompareString compares strings
func CompareString(value1 string, operator string, value2 interface{}) bool {
	value2string := fmt.Sprintf("%v", value2)
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return value1 == value2string
	case stgml.OperatorNotEqual:
		return value1 == value2string
	case stgml.OperatorRegex:
		expression := fmt.Sprintf("(?i)%s", value1)
		compiled, err := regexp.Compile(expression)
		if err != nil {
			return false
		}
		return compiled.MatchString(value2string)
	case stgml.OperatorExists:
		return value2string != ""
	case stgml.OperatorIn:
		VerifyStringSlice(value2, true, func(value string) bool {
			return value1 == value2
		})
	case stgml.OperatorNotIn:
		VerifyStringSlice(value2, false, func(value string) bool {
			return value1 != value2
		})
	}
	return false
}

// VerifyBoolSlice implementation
func VerifyBoolSlice(data interface{}, expected bool, verifyFn func(value bool) bool) bool {
	boolSlice, ok := data.([]bool)
	if !ok {
		return false
	}
	for _, value := range boolSlice {
		if verifyFn(value) == expected {
			return true
		}
	}
	return false
}

// CompareBool compares strings
func CompareBool(value1 bool, operator string, value2 interface{}) bool {
	value2bool, ok := value2.(bool)
	switch operator {
	case stgml.OperatorEqual, stgml.OperatorNone:
		return ok && value1 == value2bool
	case stgml.OperatorNotEqual:
		return ok && value1 != value2bool
	case stgml.OperatorExists:
		return ok
	case stgml.OperatorIn:
		VerifyBoolSlice(value2, true, func(value bool) bool {
			return value1 == value2
		})
	case stgml.OperatorNotIn:
		VerifyBoolSlice(value2, false, func(value bool) bool {
			return value1 != value2
		})
	}
	return false
}
