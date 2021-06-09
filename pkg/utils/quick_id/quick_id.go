package quickid

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	dataRepoML "github.com/mycontroller-org/backend/v2/pkg/model/data_repository"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	firmwareML "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	fwdPayloadML "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	gatewayML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	scheduleML "github.com/mycontroller-org/backend/v2/pkg/model/schedule"
	sourceML "github.com/mycontroller-org/backend/v2/pkg/model/source"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
)

// entity quick id prefix
var (
	QuickIDGateway        = []string{"gateway"}
	QuickIDNode           = []string{"node"}
	QuickIDSource         = []string{"source"}
	QuickIDField          = []string{"field"}
	QuickIDTask           = []string{"task"}
	QuickIDSchedule       = []string{"schedule"}
	QuickIDHandler        = []string{"handler"}
	QuickIDFirmware       = []string{"firmware"}
	QuickIDDataRepository = []string{"data_repository"}
	QuickIDForwardPayload = []string{"forward_payload"}
)

// IsValidQuickID says is it in quikID format
func IsValidQuickID(quickID string) bool {
	// get entity type and full_id
	typeID := strings.SplitN(quickID, ":", 2)
	if len(typeID) != 2 {
		return false
	}
	// entity type
	entityType := strings.ToLower(typeID[0])

	validIDs := make([]string, 0)
	validIDs = append(validIDs, QuickIDGateway...)
	validIDs = append(validIDs, QuickIDNode...)
	validIDs = append(validIDs, QuickIDSource...)
	validIDs = append(validIDs, QuickIDField...)
	validIDs = append(validIDs, QuickIDTask...)
	validIDs = append(validIDs, QuickIDSchedule...)
	validIDs = append(validIDs, QuickIDHandler...)
	validIDs = append(validIDs, QuickIDDataRepository...)
	validIDs = append(validIDs, QuickIDFirmware...)
	validIDs = append(validIDs, QuickIDForwardPayload...)

	return utils.ContainsString(validIDs, entityType)
}

// EntityKeyValueMap converts quick id to referable map
// returns entityType, keyValueMap, error
func EntityKeyValueMap(quickID string) (string, map[string]string, error) {
	data := make(map[string]string)

	// get entity type and full_id
	typeID := strings.SplitN(quickID, ":", 2)
	if len(typeID) != 2 {
		return "", nil, fmt.Errorf("invalid quick_id: %s, check the format", quickID)
	}

	// entity type
	entityType := strings.ToLower(typeID[0])

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
	case utils.ContainsString(QuickIDGateway, entityType):
		expectedLength := 1
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid gateway quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		if len(values) > expectedLength {
			data[model.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDNode, entityType):
		expectedLength := 2
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid node quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		if len(values) > expectedLength {
			data[model.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDSource, entityType):
		expectedLength := 3
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid source quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		data[model.KeySourceID] = normalizeValue(values[2])
		if len(values) > expectedLength {
			data[model.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDField, entityType):
		expectedLength := 4
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid field quick_id: %s, check the format", quickID)
		}
		data[model.KeyGatewayID] = normalizeValue(values[0])
		data[model.KeyNodeID] = normalizeValue(values[1])
		data[model.KeySourceID] = normalizeValue(values[2])
		data[model.KeyFieldID] = normalizeValue(values[3])
		if len(values) > expectedLength {
			data[model.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDTask, entityType),
		utils.ContainsString(QuickIDSchedule, entityType),
		utils.ContainsString(QuickIDHandler, entityType),
		utils.ContainsString(QuickIDDataRepository, entityType),
		utils.ContainsString(QuickIDFirmware, entityType),
		utils.ContainsString(QuickIDForwardPayload, entityType):
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
			return entityType, nil, fmt.Errorf("value missing for the field:%s, input:%s", k, quickID)
		}
	}

	return entityType, data, nil
}

// GetQuickID returns quick id of the entity
func GetQuickID(entity interface{}) (string, error) {
	itemValueType := reflect.ValueOf(entity).Kind()

	if itemValueType == reflect.Ptr {
		entity = reflect.ValueOf(entity).Elem().Interface()
		itemValueType = reflect.ValueOf(entity).Kind()
	}

	if itemValueType != reflect.Struct {
		return "", fmt.Errorf("struct type only allowed. received:%s", itemValueType.String())
	}

	itemType := reflect.TypeOf(entity)

	switch itemType {
	case reflect.TypeOf(fieldML.Field{}):
		res, ok := entity.(fieldML.Field)
		if ok {
			return fmt.Sprintf("%s:%s.%s.%s.%s", QuickIDField[0], res.GatewayID, res.NodeID, res.SourceID, res.FieldID), nil
		}

	case reflect.TypeOf(sourceML.Source{}):
		res, ok := entity.(sourceML.Source)
		if ok {
			return fmt.Sprintf("%s:%s.%s.%s", QuickIDSource[0], res.GatewayID, res.NodeID, res.SourceID), nil
		}

	case reflect.TypeOf(nodeML.Node{}):
		res, ok := entity.(nodeML.Node)
		if ok {
			return fmt.Sprintf("%s:%s.%s", QuickIDNode[0], res.GatewayID, res.NodeID), nil
		}

	case reflect.TypeOf(gatewayML.Config{}):
		res, ok := entity.(gatewayML.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDGateway[0], res.ID), nil
		}

	case reflect.TypeOf(taskML.Config{}):
		res, ok := entity.(taskML.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDTask[0], res.ID), nil
		}

	case reflect.TypeOf(scheduleML.Config{}):
		res, ok := entity.(scheduleML.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDSchedule[0], res.ID), nil
		}

	case reflect.TypeOf(handlerML.Config{}):
		res, ok := entity.(handlerML.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDHandler[0], res.ID), nil
		}

	case reflect.TypeOf(firmwareML.Firmware{}):
		res, ok := entity.(firmwareML.Firmware)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDFirmware[0], res.ID), nil
		}

	case reflect.TypeOf(dataRepoML.Config{}):
		res, ok := entity.(dataRepoML.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDDataRepository[0], res.ID), nil
		}

	case reflect.TypeOf(fwdPayloadML.Config{}):
		res, ok := entity.(fwdPayloadML.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDForwardPayload[0], res.ID), nil
		}

	default:
		return "", fmt.Errorf("unsupported type received: %s", itemType.String())
	}

	return "", errors.New("unknown resource type")
}
