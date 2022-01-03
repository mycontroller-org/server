package helper

import (
	"reflect"
	"sort"
	"time"

	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
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
				return aFinalValue < bFinalValue
			}
			return aFinalValue > bFinalValue

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
