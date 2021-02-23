package quickid

import (
	"fmt"
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
)

// resource quick id prefix
var (
	QuickIDGateway     = []string{"gw", "gateway"}
	QuickIDNode        = []string{"nd", "node"}
	QuickIDSensor      = []string{"sn", "sensor"}
	QuickIDSensorField = []string{"sf", "sensor_filed", "field"}
	QuickIDTask        = []string{"tk", "task"}
	QuickIDSchedule    = []string{"sk", "schedule"}
	QuickIDHandler     = []string{"hd", "handler"}
)

// IsValidQuickID says is it in quikID format
func IsValidQuickID(quickID string) bool {
	// get resource type and full_id
	typeID := strings.SplitN(quickID, ":", 2)
	if len(typeID) != 2 {
		return false
	}
	// resource type
	resourceType := strings.ToLower(typeID[0])

	validIDs := make([]string, 0)
	validIDs = append(validIDs, QuickIDGateway...)
	validIDs = append(validIDs, QuickIDNode...)
	validIDs = append(validIDs, QuickIDSensor...)
	validIDs = append(validIDs, QuickIDSensorField...)

	return utils.ContainsString(validIDs, resourceType)
}

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

	switch {
	case utils.ContainsString(QuickIDGateway, resourceType):
		expectedLength := 1
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("Invalid gateway quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDNode, resourceType):
		expectedLength := 2
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("Invalid node quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDSensor, resourceType):
		expectedLength := 3
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("Invalid sensor quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		data[model.KeySensorID] = normalizeValue(values[2])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDSensorField, resourceType):
		expectedLength := 4
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("Invalid sensor field quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		data[model.KeySensorID] = normalizeValue(values[2])
		data[model.KeyFieldID] = normalizeValue(values[3])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDTask, resourceType),
		utils.ContainsString(QuickIDSchedule, resourceType),
		utils.ContainsString(QuickIDHandler, resourceType):
		if typeID[1] == "" {
			return "", nil, fmt.Errorf("Invalid data. quickID:%s", quickID)
		}
		data[model.KeyID] = typeID[1]

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
