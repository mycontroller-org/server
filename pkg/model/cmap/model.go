package cmap

import (
	"fmt"
	"strconv"

	"github.com/mycontroller-org/backend/v2/pkg/utils/normalize"
	"go.uber.org/zap"
)

// CustomStringMap data
type CustomStringMap map[string]string

// Init creates if not available
func (csm CustomStringMap) Init() CustomStringMap {
	if csm == nil {
		return CustomStringMap(make(map[string]string))
	}
	return csm
}

// Clone CustomMap
func (csm CustomStringMap) Clone() CustomStringMap {
	cloned := make(map[string]string)
	for k, v := range csm {
		cloned[k] = v
	}
	return CustomStringMap(cloned)
}

// NormalizeKeys of the map
func (csm CustomStringMap) NormalizeKeys() CustomStringMap {
	newMap := make(map[string]string)
	for k, v := range csm {
		k = normalize.Key(k)
		newMap[k] = v
	}
	return newMap
}

// Set a key, value pair
func (csm CustomStringMap) Set(key, value string) {
	key = normalize.Key(key)
	m := map[string]string(csm)
	if !csm.GetBool(GetIgnoreKey(key)) { // update only if not ignored
		m[key] = value
	}
}

// Get a value by key
func (csm CustomStringMap) Get(key string) string {
	key = normalize.Key(key)
	m := map[string]string(csm)
	value, ok := m[key]
	if ok {
		return value
	}
	return ""
}

// Remove a value by key
func (csm CustomStringMap) Remove(keys ...string) {
	m := map[string]string(csm)
	for _, key := range keys {
		key = normalize.Key(key)
		delete(m, key)
	}
}

// CopyFrom another map
func (csm CustomStringMap) CopyFrom(another CustomStringMap) {
	m := map[string]string(csm)
	for k, v := range another {
		k = normalize.Key(k)
		if !csm.GetBool(GetIgnoreKey(k)) { // update only if not ignored
			m[k] = v
		}
	}
}

// GetBool a value by key
func (csm CustomStringMap) GetBool(key string) bool {
	key = normalize.Key(key)
	v, err := strconv.ParseBool(csm.Get(key))
	if err != nil {
		zap.L().Debug("error on conversion", zap.Error(err), zap.Any("value", v))
	}
	return v
}

// GetIgnoreBool a value by ignore key
func (csm CustomStringMap) GetIgnoreBool(key string) bool {
	key = normalize.Key(key)
	v, err := strconv.ParseBool(csm.Get(GetIgnoreKey(key)))
	if err != nil {
		zap.L().Debug("error on conversion", zap.Error(err), zap.Any("value", v))
	}
	return v
}

// GetInt a value by key
func (csm CustomStringMap) GetInt(key string) int {
	key = normalize.Key(key)
	v, err := strconv.ParseInt(csm.Get(key), 10, 64)
	if err != nil {
		zap.L().Debug("error on conversion", zap.Error(err), zap.Any("value", v))
	}
	return int(v)
}

// GetFloat a value by key
func (csm CustomStringMap) GetFloat(key string) float64 {
	key = normalize.Key(key)
	v, err := strconv.ParseFloat(csm.Get(key), 64)
	if err != nil {
		zap.L().Debug("error on conversion", zap.Error(err), zap.Any("value", v))
	}
	return v
}

// CustomMap data
type CustomMap map[string]interface{}

// Init creates if not available
func (cm CustomMap) Init() CustomMap {
	if cm == nil {
		return CustomMap(make(map[string]interface{}))
	}
	return cm
}

// Clone CustomMap
func (cm CustomMap) Clone() CustomMap {
	cloned := make(map[string]interface{})
	for k, v := range cm {
		cloned[k] = v
	}
	return CustomMap(cloned)
}

// NormalizeKeys of the map
func (cm CustomMap) NormalizeKeys() CustomMap {
	newMap := make(map[string]interface{})
	for k, v := range cm {
		k = normalize.Key(k)
		newMap[k] = v
	}
	return newMap
}

// ToMap returns as map
func (cm CustomMap) ToMap() map[string]interface{} {
	return map[string]interface{}(cm)
}

// Set a key, value pair
func (cm CustomMap) Set(key string, value interface{}, labels CustomStringMap) {
	key = normalize.Key(key)
	if labels != nil {
		if labels.GetBool(GetIgnoreKey(key)) { // if ignored, do not update
			return
		}
	}
	// update value
	m := map[string]interface{}(cm)
	m[key] = value
}

// Remove a key
func (cm CustomMap) Remove(keys ...string) {
	m := map[string]interface{}(cm)
	for _, key := range keys {
		key = normalize.Key(key)
		delete(m, key)
	}
}

// CopyFrom another map
func (cm CustomMap) CopyFrom(another CustomMap, labels CustomStringMap) {
	m := map[string]interface{}(cm)
	for k, v := range another {
		k = normalize.Key(k)
		if labels != nil {
			if labels.GetBool(GetIgnoreKey(k)) { // if ignored, do not update
				continue
			}
		}
		m[k] = v
	}
}

// Get a value by key
func (cm CustomMap) Get(key string) interface{} {
	key = normalize.Key(key)
	m := map[string]interface{}(cm)
	for k, v := range m {
		k = normalize.Key(k)
		if key == k {
			return v
		}
	}
	return nil
}

// GetString a value by key
func (cm CustomMap) GetString(key string) string {
	key = normalize.Key(key)
	originalValue := cm.Get(key)
	if originalValue == nil {
		return ""
	}
	finalValue, ok := originalValue.(string)
	if ok {
		return finalValue
	}
	return fmt.Sprintf("%v", originalValue)
}

// GetBool a value by key
func (cm CustomMap) GetBool(key string) bool {
	key = normalize.Key(key)
	originalValue := cm.Get(key)
	if originalValue == nil {
		return false
	}

	finalValue, ok := originalValue.(bool)
	if ok {
		return finalValue
	}

	finalValue, err := strconv.ParseBool(fmt.Sprintf("%v", originalValue))
	if err != nil {
		return false
	}
	return finalValue
}

// GetInt64 a value by key
func (cm CustomMap) GetInt64(key string) int64 {
	key = normalize.Key(key)
	originalValue := cm.Get(key)
	if originalValue == nil {
		return 0
	}
	finalValue, ok := originalValue.(int64)
	if ok {
		return finalValue
	}
	finalValue, err := strconv.ParseInt(fmt.Sprintf("%v", originalValue), 10, 64)
	if err != nil {
		return 0
	}
	return finalValue
}

// GetFloat64 a value by key
func (cm CustomMap) GetFloat64(key string) float64 {
	key = normalize.Key(key)
	originalValue := cm.Get(key)
	if originalValue == nil {
		return 0
	}
	finalValue, ok := originalValue.(float64)
	if ok {
		return finalValue
	}
	finalValue, err := strconv.ParseFloat(fmt.Sprintf("%v", originalValue), 64)
	if err != nil {
		return 0
	}
	return finalValue
}

// GetIgnoreKey of a key
func GetIgnoreKey(key string) string {
	return fmt.Sprintf("ignore_%s", key)
}
