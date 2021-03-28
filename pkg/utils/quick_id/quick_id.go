package quickid

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/field"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
)

// resource quick id prefix
var (
	QuickIDGateway        = []string{"gw", "gateway"}
	QuickIDNode           = []string{"nd", "node"}
	QuickIDSource         = []string{"sn", "source"}
	QuickIDField          = []string{"sf", "field"}
	QuickIDTask           = []string{"tk", "task"}
	QuickIDSchedule       = []string{"sk", "schedule"}
	QuickIDHandler        = []string{"hd", "handler"}
	QuickIDDataRepository = []string{"dr", "data_repository"}
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
	validIDs = append(validIDs, QuickIDSource...)
	validIDs = append(validIDs, QuickIDField...)
	validIDs = append(validIDs, QuickIDTask...)
	validIDs = append(validIDs, QuickIDSchedule...)
	validIDs = append(validIDs, QuickIDHandler...)
	validIDs = append(validIDs, QuickIDDataRepository...)

	return utils.ContainsString(validIDs, resourceType)
}

// ResourceKeyValueMap converts quick id to referable map
// returns resourceType, keyValueMap, error
func ResourceKeyValueMap(quickID string) (string, map[string]string, error) {
	data := make(map[string]string)

	// get resource type and full_id
	typeID := strings.SplitN(quickID, ":", 2)
	if len(typeID) != 2 {
		return "", nil, fmt.Errorf("invalid quick_id: %s, check the format", quickID)
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
			return "", nil, fmt.Errorf("invalid gateway quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDNode, resourceType):
		expectedLength := 2
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid node quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDSource, resourceType):
		expectedLength := 3
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid source quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		data[model.KeySourceID] = normalizeValue(values[2])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDField, resourceType):
		expectedLength := 4
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid field quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		data[model.KeySourceID] = normalizeValue(values[2])
		data[model.KeyFieldID] = normalizeValue(values[3])
		if len(values) > expectedLength {
			data[model.KeySelector] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDTask, resourceType),
		utils.ContainsString(QuickIDSchedule, resourceType),
		utils.ContainsString(QuickIDHandler, resourceType),
		utils.ContainsString(QuickIDDataRepository, resourceType):
		if typeID[1] == "" {
			return "", nil, fmt.Errorf("invalid data. quickID:%s", quickID)
		}
		data[model.KeyID] = typeID[1]

	default:
		return "", nil, fmt.Errorf("invalid resource type quick_id: %s, check the format", quickID)
	}

	// verify all the fields are available
	for k, v := range data {
		if v == "" {
			return resourceType, nil, fmt.Errorf("value missing for the field:%s, input:%s", k, quickID)
		}
	}

	return resourceType, data, nil
}

// GetQuickID returns quick id of the resource
func GetQuickID(item interface{}) (string, error) {
	itemValueType := reflect.ValueOf(item).Kind()
	if itemValueType != reflect.Struct {
		return "", fmt.Errorf("struct type only allowed. received:%s", itemValueType.String())
	}

	itemType := reflect.TypeOf(item)

	switch itemType {
	case reflect.TypeOf(fieldML.Field{}):
		field := item.(field.Field)
		return fmt.Sprintf("%s:%s.%s.%s.%s", QuickIDField[1], field.GatewayID, field.NodeID, field.SourceID, field.FieldID), nil

	default:
		return "", fmt.Errorf("unsupported struct. received:%s", itemType.String())
	}
}
