package apply

import "fmt"

var arrayMergeKeys = []string{"id", "key", "fieldId", "field", "name", "type", "sourceId"}

func deepMergeValue(base, overlay interface{}) interface{} {
	if overlay == nil {
		return base
	}
	baseMap, baseIsMap := asMap(base)
	overlayMap, overlayIsMap := asMap(overlay)
	if baseIsMap && overlayIsMap {
		return deepMergeMaps(baseMap, overlayMap)
	}
	baseSlice, baseIsSlice := asSlice(base)
	overlaySlice, overlayIsSlice := asSlice(overlay)
	if overlayIsSlice {
		if baseIsSlice {
			return deepMergeSlices(baseSlice, overlaySlice)
		}
		return overlaySlice
	}
	return overlay
}

func deepMergeMaps(base, overlay map[string]interface{}) map[string]interface{} {
	out := copyMap(base)
	for key, value := range overlay {
		if existing, ok := out[key]; ok {
			out[key] = deepMergeValue(existing, value)
			continue
		}
		out[key] = value
	}
	return out
}

func deepMergeSlices(base, overlay []interface{}) []interface{} {
	key := arrayItemKey(overlay)
	if key == "" {
		key = arrayItemKey(base)
	}
	if key == "" {
		return overlay
	}

	out := make([]interface{}, len(base))
	copy(out, base)
	index := map[string]int{}
	for i, item := range out {
		if m, ok := asMap(item); ok {
			if id := mapString(m, key); id != "" {
				index[id] = i
			}
		}
	}
	for _, item := range overlay {
		m, ok := asMap(item)
		if !ok {
			out = append(out, item)
			continue
		}
		id := mapString(m, key)
		if id == "" {
			out = append(out, item)
			continue
		}
		if pos, found := index[id]; found {
			out[pos] = deepMergeValue(out[pos], m)
			continue
		}
		index[id] = len(out)
		out = append(out, item)
	}
	return out
}

func arrayItemKey(items []interface{}) string {
	for _, item := range items {
		m, ok := asMap(item)
		if !ok {
			continue
		}
		for _, key := range arrayMergeKeys {
			if mapString(m, key) != "" {
				return key
			}
		}
	}
	return ""
}

func asMap(value interface{}) (map[string]interface{}, bool) {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed, true
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			out[fmt.Sprint(key)] = item
		}
		return out, true
	default:
		return nil, false
	}
}

func asSlice(value interface{}) ([]interface{}, bool) {
	switch typed := value.(type) {
	case []interface{}:
		return typed, true
	default:
		return nil, false
	}
}

func mapString(m map[string]interface{}, key string) string {
	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}
	s := fmt.Sprint(value)
	if s == "" || s == "<nil>" {
		return ""
	}
	return s
}
