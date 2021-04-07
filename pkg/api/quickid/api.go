package quickid

import (
	"fmt"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/notify_handler"
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	sourceAPI "github.com/mycontroller-org/backend/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	quickIdUtils "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
)

// GetResources returns resource
func GetResources(quickIDs []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, quickID := range quickIDs {
		resourceType, keys, err := quickIdUtils.ResourceKeyValueMap(quickID)
		if err != nil {
			return result, err
		}

		var item interface{}

		switch resourceType {
		case "gateway":
			item, err = gatewayAPI.GetByID(keys[model.KeyGatewayID])

		case "node":
			item, err = nodeAPI.GetByGatewayAndNodeID(keys[model.KeyGatewayID], keys[model.KeyNodeID])

		case "source":
			item, err = sourceAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySourceID])

		case "field":
			item, err = fieldAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySourceID], keys[model.KeyFieldID])

		case "task":
			item, err = taskAPI.GetByID(keys[model.KeyID])

		case "schedule":
			item, err = schedulerAPI.GetByID(keys[model.KeyID])

		case "handler":
			item, err = handlerAPI.GetByID(keys[model.KeyID])

		default:
			return nil, fmt.Errorf("unknown resource type: %s, quickID: %s", resourceType, quickID)
		}

		if err != nil {
			return nil, err
		}

		if item != nil {
			result[quickID] = item
		}

	}
	return result, nil
}
