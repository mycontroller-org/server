package quickid

import (
	"fmt"

	dataRepoAPI "github.com/mycontroller-org/server/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	handlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	sourceAPI "github.com/mycontroller-org/server/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
)

// GetResources returns resource
func GetResources(quickIDs []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, quickID := range quickIDs {
		resourceType, keys, err := quickIdUtils.EntityKeyValueMap(quickID)
		if err != nil {
			return result, err
		}

		var item interface{}

		switch resourceType {
		case types.EntityGateway:
			item, err = gatewayAPI.GetByID(keys[types.KeyGatewayID])

		case types.EntityNode:
			item, err = nodeAPI.GetByGatewayAndNodeID(keys[types.KeyGatewayID], keys[types.KeyNodeID])

		case types.EntitySource:
			item, err = sourceAPI.GetByIDs(keys[types.KeyGatewayID], keys[types.KeyNodeID], keys[types.KeySourceID])

		case types.EntityField:
			item, err = fieldAPI.GetByIDs(keys[types.KeyGatewayID], keys[types.KeyNodeID], keys[types.KeySourceID], keys[types.KeyFieldID])

		case types.EntityTask:
			item, err = taskAPI.GetByID(keys[types.KeyID])

		case types.EntitySchedule:
			item, err = scheduleAPI.GetByID(keys[types.KeyID])

		case types.EntityHandler:
			item, err = handlerAPI.GetByID(keys[types.KeyID])

		case types.EntityDataRepository:
			item, err = dataRepoAPI.GetByID(keys[types.KeyID])

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
