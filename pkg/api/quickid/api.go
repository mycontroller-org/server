package quickid

import (
	"fmt"

	dataRepoAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/backend/v2/pkg/api/schedule"
	sourceAPI "github.com/mycontroller-org/backend/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	quickIdUtils "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
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
		case model.EntityGateway:
			item, err = gatewayAPI.GetByID(keys[model.KeyGatewayID])

		case model.EntityNode:
			item, err = nodeAPI.GetByGatewayAndNodeID(keys[model.KeyGatewayID], keys[model.KeyNodeID])

		case model.EntitySource:
			item, err = sourceAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySourceID])

		case model.EntityField:
			item, err = fieldAPI.GetByIDs(keys[model.KeyGatewayID], keys[model.KeyNodeID], keys[model.KeySourceID], keys[model.KeyFieldID])

		case model.EntityTask:
			item, err = taskAPI.GetByID(keys[model.KeyID])

		case model.EntitySchedule:
			item, err = scheduleAPI.GetByID(keys[model.KeyID])

		case model.EntityHandler:
			item, err = handlerAPI.GetByID(keys[model.KeyID])

		case model.EntityDataRepository:
			item, err = dataRepoAPI.GetByID(keys[model.KeyID])

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
