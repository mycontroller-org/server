package model

import (
	"fmt"
	"strings"
)

// resource quick id prefix
const (
	QuickIDGateway     = "gw"
	QuickIDNode        = "no"
	QuickIDNodeData    = "nd"
	QuickIDSensor      = "sn"
	QuickIDSensorField = "sf"
)

// ResourceKeyValueMap converts quick id to referable map
// returns resourceType, keyValueMap, error
func ResourceKeyValueMap(quickID string) (string, map[string]string, error) {
	data := make(map[string]string)

	// get resource type and full_id
	typeID := strings.SplitN(quickID, ":", 2)
	if len(typeID) != 2 {
		return "", nil, fmt.Errorf("Invalid quick_id: %s, check the format", quickID)
	}

	// resource type
	resourceType := strings.ToLower(typeID[0])

	// id values
	values := strings.Split(typeID[1], ".")

	normalizeValue := func(value string) string {
		v := strings.TrimSpace(value)
		if len(v) > 0 {
			return v
		}
		return ""
	}

	switch resourceType {
	case QuickIDGateway:
		if len(values) != 1 {
			return "", nil, fmt.Errorf("Invalid gateway quick_id: %s, check the format", quickID)
		}
		data[KeyGatewayID] = normalizeValue(values[0])

	case QuickIDNode:
		if len(values) != 2 {
			return "", nil, fmt.Errorf("Invalid node quick_id: %s, check the format", quickID)
		}
		data[KeyGatewayID] = normalizeValue(values[0])
		data[KeyNodeID] = normalizeValue(values[1])

	case QuickIDNodeData:
		if len(values) != 3 {
			return "", nil, fmt.Errorf("Invalid node data quick_id: %s, check the format", quickID)
		}
		data[KeyGatewayID] = normalizeValue(values[0])
		data[KeyNodeID] = normalizeValue(values[1])
		data[KeyFieldID] = normalizeValue(values[2])

	case QuickIDSensor:
		if len(values) != 3 {
			return "", nil, fmt.Errorf("Invalid sensor quick_id: %s, check the format", quickID)
		}
		data[KeyGatewayID] = normalizeValue(values[0])
		data[KeyNodeID] = normalizeValue(values[1])
		data[KeySensorID] = normalizeValue(values[2])

	case QuickIDSensorField:
		if len(values) != 4 {
			return "", nil, fmt.Errorf("Invalid sensor field quick_id: %s, check the format", quickID)
		}
		data[KeyGatewayID] = normalizeValue(values[0])
		data[KeyNodeID] = normalizeValue(values[1])
		data[KeySensorID] = normalizeValue(values[2])
		data[KeyFieldID] = normalizeValue(values[3])

	default:
		return "", nil, fmt.Errorf("Invalid resource type quick_id: %s, check the format", quickID)
	}

	// verify all the fields are available
	for k, v := range data {
		if v == "" {
			return resourceType, nil, fmt.Errorf("Value missing for the field:%s, input:%s", k, quickID)
		}
	}

	return resourceType, data, nil
}
