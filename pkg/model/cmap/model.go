package cmap

import (
	"fmt"
	"strconv"
	"strings"
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

// Set a key, value pair
func (csm CustomStringMap) Set(key, value string) {
	m := map[string]string(csm)
	if !csm.GetBool(GetIgnoreKey(key)) { // update only if not ignored
		m[strings.ToLower(key)] = value
	}
}

// Get a value by key
func (csm CustomStringMap) Get(key string) string {
	m := map[string]string(csm)
	value, ok := m[strings.ToLower(key)]
	if ok {
		return value
	}
	return ""
}

// Remove a value by key
func (csm CustomStringMap) Remove(keys ...string) {
	m := map[string]string(csm)
	for _, key := range keys {
		delete(m, strings.ToLower(key))
	}
}

// CopyFrom another map
func (csm CustomStringMap) CopyFrom(another CustomStringMap) {
	m := map[string]string(csm)
	for k, v := range another {
		if !csm.GetBool(GetIgnoreKey(k)) { // update only if not ignored
			m[strings.ToLower(k)] = v
		}
	}
}

// GetBool a value by key
func (csm CustomStringMap) GetBool(key string) bool {
	v, err := strconv.ParseBool(strings.ToLower(csm.Get(key)))
	if err != nil {
		// TODO: needs to pass it to logger?
	}
	return v
}

// GetIgnoreBool a value by ignore key
func (csm CustomStringMap) GetIgnoreBool(key string) bool {
	v, err := strconv.ParseBool(strings.ToLower(csm.Get(GetIgnoreKey(key))))
	if err != nil {
		// TODO: needs to pass it to logger?
	}
	return v
}

// GetInt a value by key
func (csm CustomStringMap) GetInt(key string) int {
	v, err := strconv.ParseInt(csm.Get(key), 10, 64)
	if err != nil {
		// TODO: needs to pass it to logger?
	}
	return int(v)
}

// GetFloat a value by key
func (csm CustomStringMap) GetFloat(key string) float64 {
	v, err := strconv.ParseFloat(csm.Get(key), 64)
	if err != nil {
		// TODO: needs to pass it to logger?
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

// Set a key, value pair
func (cm CustomMap) Set(key string, value interface{}, labels CustomStringMap) {
	if labels != nil {
		if labels.GetBool(GetIgnoreKey(key)) { // if ignored, do not update
			return
		}
	}
	// update value
	m := map[string]interface{}(cm)
	m[strings.ToLower(key)] = value
}

// Remove a key
func (cm CustomMap) Remove(keys ...string) {
	m := map[string]interface{}(cm)
	for _, key := range keys {
		delete(m, strings.ToLower(key))
	}
}

// CopyFrom another map
func (cm CustomMap) CopyFrom(another CustomMap, labels CustomStringMap) {
	m := map[string]interface{}(cm)
	for k, v := range another {
		if labels != nil {
			if labels.GetBool(GetIgnoreKey(k)) { // if ignored, do not update
				continue
			}
		}
		m[strings.ToLower(k)] = v
	}
}

// Get a value by key
func (cm CustomMap) Get(key string) interface{} {
	m := map[string]interface{}(cm)
	value, ok := m[strings.ToLower(key)]
	if ok {
		return value
	}
	return ""
}

// GetIgnoreKey of a key
func GetIgnoreKey(key string) string {
	return fmt.Sprintf("ignore_%s", key)
}
