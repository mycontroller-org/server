package helper

import (
	"reflect"
	"sort"
	"time"
	"unicode"

	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// Sort given slice
func Sort(entities []interface{}, pagination *storageTY.Pagination) ([]interface{}, int64) {
	entitiesCount := int64(len(entities))
	if pagination == nil {
		return entities, entitiesCount
	}

	if entitiesCount == 0 {
		return entities, entitiesCount
	}
	// sort entities
	if len(pagination.SortBy) > 0 {
		// supports only one sort option, take the first one
		s := pagination.SortBy[0]
		entities = GetSortByKeyPath(s.Field, s.OrderBy, entities)
	}

	if pagination.Limit > 0 {
		posStart := pagination.Offset
		posEnd := posStart + pagination.Limit
		if entitiesCount > posEnd {
			return entities[posStart:posEnd], entitiesCount
		} else if entitiesCount > posStart {
			return entities[posStart:], entitiesCount
		} else {
			return entities[:], entitiesCount
		}
	}

	return entities, entitiesCount
}

// naturalStringLess compares strings using natural sort order
// Numbers within strings are compared numerically
func naturalStringLess(a, b string) bool {
	aRunes := []rune(a)
	bRunes := []rune(b)
	i, j := 0, 0

	for i < len(aRunes) && j < len(bRunes) {
		// Check if both are digits
		if unicode.IsDigit(aRunes[i]) && unicode.IsDigit(bRunes[j]) {
			// Extract numbers
			aNum, aEnd := extractNumber(aRunes, i)
			bNum, bEnd := extractNumber(bRunes, j)

			if aNum != bNum {
				return aNum < bNum
			}
			i = aEnd
			j = bEnd
		} else {
			if aRunes[i] != bRunes[j] {
				return aRunes[i] < bRunes[j]
			}
			i++
			j++
		}
	}
	return len(aRunes) < len(bRunes)
}

// extractNumber extracts a number from rune slice starting at pos
func extractNumber(runes []rune, pos int) (int, int) {
	num := 0
	for pos < len(runes) && unicode.IsDigit(runes[pos]) {
		num = num*10 + int(runes[pos]-'0')
		pos++
	}
	return num, pos
}

// GetSortByKeyPath returns the slice in order
func GetSortByKeyPath(keyPath, orderBy string, data []interface{}) []interface{} {
	sort.Slice(data, func(a, b int) bool {
		aKind, aValue, err := GetValueByKeyPath(data[a], keyPath)
		if err != nil {
			return false
		}
		bKind, bValue, err := GetValueByKeyPath(data[b], keyPath)
		if err != nil {
			return false
		}

		if aKind != bKind {
			return false
		}

		switch aKind {
		case reflect.String:
			aFinalValue, aOK := aValue.(string)
			bFinalValue, bOK := bValue.(string)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return naturalStringLess(aFinalValue, bFinalValue)
			}
			return naturalStringLess(bFinalValue, aFinalValue)

		case reflect.Int:
			aFinalValue, aOK := aValue.(int)
			bFinalValue, bOK := bValue.(int)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Int8:
			aFinalValue, aOK := aValue.(int8)
			bFinalValue, bOK := bValue.(int8)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Int16:
			aFinalValue, aOK := aValue.(int16)
			bFinalValue, bOK := bValue.(int16)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Int32:
			aFinalValue, aOK := aValue.(int32)
			bFinalValue, bOK := bValue.(int32)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Int64:
			aFinalValue, aOK := aValue.(int64)
			bFinalValue, bOK := bValue.(int64)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Float32:
			aFinalValue, aOK := aValue.(float32)
			bFinalValue, bOK := bValue.(float32)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Float64:
			aFinalValue, aOK := aValue.(float64)
			bFinalValue, bOK := bValue.(float64)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Bool:
			aFinalValue, aOK := aValue.(bool)
			bFinalValue, bOK := bValue.(bool)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue
			}
			return !bFinalValue

		case reflect.Uint:
			aFinalValue, aOK := aValue.(uint)
			bFinalValue, bOK := bValue.(uint)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Uint8:
			aFinalValue, aOK := aValue.(uint8)
			bFinalValue, bOK := bValue.(uint8)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Uint16:
			aFinalValue, aOK := aValue.(uint16)
			bFinalValue, bOK := bValue.(uint16)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Uint32:
			aFinalValue, aOK := aValue.(uint32)
			bFinalValue, bOK := bValue.(uint32)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Uint64:
			aFinalValue, aOK := aValue.(uint64)
			bFinalValue, bOK := bValue.(uint64)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

		case reflect.Struct:
			aFinalValue, aOK := aValue.(time.Time)
			bFinalValue, bOK := bValue.(time.Time)
			if !aOK || !bOK {
				return false
			}
			if orderBy == storageTY.SortByASC {
				return aFinalValue.After(bFinalValue)
			}
			return aFinalValue.Before(bFinalValue)
		}
		return false
	})
	return data
}
