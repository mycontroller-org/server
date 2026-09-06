package helper

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// FilterByStringMap filter
func FilterByStringMap(entities []interface{}, filtersMap map[string]string) []interface{} {
	filters := make([]storageTY.Filter, 0)
	for k, v := range filtersMap {
		filters = append(filters, storageTY.Filter{Key: k, Operator: storageTY.OperatorEqual, Value: v})
	}
	if len(filters) == 0 {
		return entities
	}
	return Filter(entities, filters, false)
}

// Filter filters the given slice
func Filter(entities []interface{}, filters []storageTY.Filter, returnSingle bool) []interface{} {
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
func IsMatching(entity interface{}, filters []storageTY.Filter) bool {
	match := true
	for index := 0; index < len(filters); index++ {
		filter := filters[index]

		// OR of filter groups (used by RBAC list scoping)
		if filter.Operator == storageTY.OperatorOr {
			if !matchOrGroups(entity, filter.Value) {
				match = false
				break
			}
			continue
		}

		// NOR of one AND-group (RBAC Deny exclude: NOT (gw=X AND node=Y))
		if filter.Operator == storageTY.OperatorNor {
			if matchNorGroup(entity, filter.Value) {
				// matchNorGroup true means entity IS in the denied set
				match = false
				break
			}
			continue
		}

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

		case reflect.Struct:
			timeValue, ok := value.(time.Time)
			if !ok {
				match = false
				break
			}
			match = CompareTime(timeValue, filter.Operator, filter.Value)

		default:
			match = false
		}

		if !match {
			break
		}
	}
	return match
}

// matchOrGroups evaluates Value as [][]Filter (OR of AND-groups).
func matchOrGroups(entity interface{}, value interface{}) bool {
	groups, ok := value.([][]storageTY.Filter)
	if !ok {
		// tolerate []interface{} from generic unmarshalling
		raw, ok := value.([]interface{})
		if !ok {
			return false
		}
		groups = make([][]storageTY.Filter, 0, len(raw))
		for _, g := range raw {
			switch gg := g.(type) {
			case []storageTY.Filter:
				groups = append(groups, gg)
			case []interface{}:
				inner := make([]storageTY.Filter, 0, len(gg))
				for _, item := range gg {
					if f, ok := item.(storageTY.Filter); ok {
						inner = append(inner, f)
					}
				}
				groups = append(groups, inner)
			}
		}
	}
	if len(groups) == 0 {
		return false
	}
	for _, group := range groups {
		if IsMatching(entity, group) {
			return true
		}
	}
	return false
}

// matchNorGroup returns true if the entity matches the positive AND-group
// (i.e. should be excluded by OperatorNor). Value is []Filter.
func matchNorGroup(entity interface{}, value interface{}) bool {
	group, ok := value.([]storageTY.Filter)
	if !ok {
		raw, ok := value.([]interface{})
		if !ok {
			return false
		}
		group = make([]storageTY.Filter, 0, len(raw))
		for _, item := range raw {
			if f, ok := item.(storageTY.Filter); ok {
				group = append(group, f)
			}
		}
	}
	if len(group) == 0 {
		return false
	}
	return IsMatching(entity, group)
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
	case storageTY.OperatorIn:
		for _, fValue := range stringSlice {
			if value == fValue {
				return true
			}
		}
	case storageTY.OperatorNotIn:
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
	case storageTY.OperatorEqual, storageTY.OperatorNone:
		return converterUtils.ToString(filterValue) == valueString
	case storageTY.OperatorNotEqual:
		return converterUtils.ToString(filterValue) != valueString
	case storageTY.OperatorRegex:
		expression := fmt.Sprintf("(?i)%s", converterUtils.ToString(filterValue))
		compiled, err := regexp.Compile(expression)
		if err != nil {
			return false
		}
		return compiled.MatchString(valueString)
	case storageTY.OperatorRegexCaseSensitive:
		compiled, err := regexp.Compile(converterUtils.ToString(filterValue))
		if err != nil {
			return false
		}
		return compiled.MatchString(valueString)
	case storageTY.OperatorExists:
		return valueString != ""
	case storageTY.OperatorIn, storageTY.OperatorNotIn:
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
	case storageTY.OperatorIn:
		for _, fValue := range boolSlice {
			if value == fValue {
				return true
			}
		}
	case storageTY.OperatorNotIn:
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
	case storageTY.OperatorEqual, storageTY.OperatorNone:
		return converterUtils.ToBool(value) == converterUtils.ToBool(expectedValue)
	case storageTY.OperatorNotEqual:
		return converterUtils.ToBool(value) != converterUtils.ToBool(expectedValue)
	case storageTY.OperatorExists:
		return len(converterUtils.ToString(value)) > 0
	case storageTY.OperatorIn, storageTY.OperatorNotIn:
		return VerifyBoolSlice(converterUtils.ToBool(value), operator, expectedValue)
	}
	return false
}

// CompareFloat compares float
func CompareFloat(value interface{}, operator string, expectedValue interface{}) bool {
	valueFloat := converterUtils.ToFloat(value)
	switch operator {
	case storageTY.OperatorEqual, storageTY.OperatorNone:
		return valueFloat == converterUtils.ToFloat(expectedValue)

	case storageTY.OperatorNotEqual:
		return valueFloat != converterUtils.ToFloat(expectedValue)

	case storageTY.OperatorGreaterThan:
		return valueFloat > converterUtils.ToFloat(expectedValue)

	case storageTY.OperatorGreaterThanEqual:
		return valueFloat >= converterUtils.ToFloat(expectedValue)

	case storageTY.OperatorLessThan:
		return valueFloat < converterUtils.ToFloat(expectedValue)

	case storageTY.OperatorLessThanEqual:
		return valueFloat <= converterUtils.ToFloat(expectedValue)

	case storageTY.OperatorIn, storageTY.OperatorNotIn, storageTY.OperatorRangeIn, storageTY.OperatorRangeNotIn:
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
	case storageTY.OperatorIn:
		for _, fValue := range floatSlice {
			if value == fValue {
				return true
			}
		}

	case storageTY.OperatorNotIn:
		for _, fValue := range floatSlice {
			if value == fValue {
				return false
			}
		}
		return true

	case storageTY.OperatorRangeIn:
		if len(floatSlice) != 2 {
			return false
		}
		lowRange := floatSlice[0]
		highRange := floatSlice[1]
		return value > lowRange && value < highRange

	case storageTY.OperatorRangeNotIn:
		if len(floatSlice) != 2 {
			return false
		}
		lowRange := floatSlice[0]
		highRange := floatSlice[1]
		return value < lowRange || value > highRange

	}
	return false
}

// CompareTime compares time.Time values
func CompareTime(value time.Time, operator string, expectedValue interface{}) bool {
	expectedStr := converterUtils.ToString(expectedValue)
	expected, err := time.Parse(time.RFC3339, expectedStr)
	if err != nil {
		return false
	}
	switch operator {
	case storageTY.OperatorEqual, storageTY.OperatorNone:
		return value.Equal(expected)
	case storageTY.OperatorNotEqual:
		return !value.Equal(expected)
	case storageTY.OperatorGreaterThan:
		return value.After(expected)
	case storageTY.OperatorGreaterThanEqual:
		return value.After(expected) || value.Equal(expected)
	case storageTY.OperatorLessThan:
		return value.Before(expected)
	case storageTY.OperatorLessThanEqual:
		return value.Before(expected) || value.Equal(expected)
	}
	return false
}

// IsMine verifies the supplied id and labels with valid list
func IsMine(svcFilter *sfTY.ServiceFilter, targetType, targetID string, targetLabels cmap.CustomStringMap) bool {
	if !svcFilter.HasFilter() {
		return true
	}

	matches := make([]bool, 0)

	if len(svcFilter.Types) > 0 {
		matched := false
		for _, typeString := range svcFilter.Types {
			if typeString == targetType {
				if !svcFilter.MatchAll {
					return true
				}
				matched = true
				break
			}
		}
		matches = append(matches, matched)
	}

	if len(svcFilter.IDs) > 0 {
		matched := false
		for _, id := range svcFilter.IDs {
			if id == targetID {
				if !svcFilter.MatchAll {
					return true
				}
				matched = true
				break
			}
		}
		matches = append(matches, matched)
	}

	if len(svcFilter.Labels) > 0 {
		matched := true
		for key, value := range svcFilter.Labels {
			receivedValue := targetLabels.Get(key)
			if value != receivedValue {
				if !svcFilter.MatchAll {
					return false
				}
				matched = false
				break
			}
		}
		matches = append(matches, matched)
	}

	if svcFilter.MatchAll {
		for _, matched := range matches {
			if !matched {
				return false
			}
		}
		return true
	}

	for _, matched := range matches {
		if matched {
			return true
		}
	}

	return false
}
