package quickid

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/types"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
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
		data[types.KeyGatewayID] = normalizeValue(values[0])
		if len(values) > expectedLength {
			data[types.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDNode, entityType):
		expectedLength := 2
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid node quick_id: %s, check the format", quickID)
		}
		data[types.KeyGatewayID] = normalizeValue(values[0])
		data[types.KeyNodeID] = normalizeValue(values[1])
		if len(values) > expectedLength {
			data[types.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDSource, entityType):
		expectedLength := 3
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid source quick_id: %s, check the format", quickID)
		}
		data[types.KeyGatewayID] = normalizeValue(values[0])
		data[types.KeyNodeID] = normalizeValue(values[1])
		data[types.KeySourceID] = normalizeValue(values[2])
		if len(values) > expectedLength {
			data[types.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
		}

	case utils.ContainsString(QuickIDField, entityType):
		expectedLength := 4
		if len(values) < expectedLength {
			return "", nil, fmt.Errorf("invalid field quick_id: %s, check the format", quickID)
		}
		data[types.KeyGatewayID] = normalizeValue(values[0])
		data[types.KeyNodeID] = normalizeValue(values[1])
		data[types.KeySourceID] = normalizeValue(values[2])
		data[types.KeyFieldID] = normalizeValue(values[3])
		if len(values) > expectedLength {
			data[types.KeyKeyPath] = normalizeValue(strings.Join(values[expectedLength:], "."))
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
		data[types.KeyID] = typeID[1]

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
	case reflect.TypeOf(fieldTY.Field{}):
		res, ok := entity.(fieldTY.Field)
		if ok {
			return fmt.Sprintf("%s:%s.%s.%s.%s", QuickIDField[0], res.GatewayID, res.NodeID, res.SourceID, res.FieldID), nil
		}

	case reflect.TypeOf(sourceTY.Source{}):
		res, ok := entity.(sourceTY.Source)
		if ok {
			return fmt.Sprintf("%s:%s.%s.%s", QuickIDSource[0], res.GatewayID, res.NodeID, res.SourceID), nil
		}

	case reflect.TypeOf(nodeTY.Node{}):
		res, ok := entity.(nodeTY.Node)
		if ok {
			return fmt.Sprintf("%s:%s.%s", QuickIDNode[0], res.GatewayID, res.NodeID), nil
		}

	case reflect.TypeOf(gatewayTY.Config{}):
		res, ok := entity.(gatewayTY.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDGateway[0], res.ID), nil
		}

	case reflect.TypeOf(taskTY.Config{}):
		res, ok := entity.(taskTY.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDTask[0], res.ID), nil
		}

	case reflect.TypeOf(scheduleTY.Config{}):
		res, ok := entity.(scheduleTY.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDSchedule[0], res.ID), nil
		}

	case reflect.TypeOf(handlerTY.Config{}):
		res, ok := entity.(handlerTY.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDHandler[0], res.ID), nil
		}

	case reflect.TypeOf(firmwareTY.Firmware{}):
		res, ok := entity.(firmwareTY.Firmware)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDFirmware[0], res.ID), nil
		}

	case reflect.TypeOf(dataRepoTY.Config{}):
		res, ok := entity.(dataRepoTY.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDDataRepository[0], res.ID), nil
		}

	case reflect.TypeOf(fwdPayloadTY.Config{}):
		res, ok := entity.(fwdPayloadTY.Config)
		if ok {
			return fmt.Sprintf("%s:%s", QuickIDForwardPayload[0], res.ID), nil
		}

	default:
		return "", fmt.Errorf("unsupported type received: %s", itemType.String())
	}

	return "", errors.New("unknown resource type")
}
