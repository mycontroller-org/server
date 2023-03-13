package quickid

import (
	"context"
	"fmt"

	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
)

type QuickIdAPI struct {
	api *entityAPI.API
}

func New(ctx context.Context) (*QuickIdAPI, error) {
	entityAPI, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &QuickIdAPI{
		api: entityAPI,
	}, nil
}

// GetResources returns resource
func (qi *QuickIdAPI) GetResources(quickIDs []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, quickID := range quickIDs {
		resourceType, keys, err := quickIdUtils.EntityKeyValueMap(quickID)
		if err != nil {
			return result, err
		}

		var item interface{}

		switch resourceType {
		case types.EntityGateway:
			item, err = qi.api.Gateway().GetByID(keys[types.KeyGatewayID])

		case types.EntityNode:
			item, err = qi.api.Node().GetByGatewayAndNodeID(keys[types.KeyGatewayID], keys[types.KeyNodeID])

		case types.EntitySource:
			item, err = qi.api.Source().GetByIDs(keys[types.KeyGatewayID], keys[types.KeyNodeID], keys[types.KeySourceID])

		case types.EntityField:
			item, err = qi.api.Field().GetByIDs(keys[types.KeyGatewayID], keys[types.KeyNodeID], keys[types.KeySourceID], keys[types.KeyFieldID])

		case types.EntityTask:
			item, err = qi.api.Task().GetByID(keys[types.KeyID])

		case types.EntitySchedule:
			item, err = qi.api.Schedule().GetByID(keys[types.KeyID])

		case types.EntityHandler:
			item, err = qi.api.Handler().GetByID(keys[types.KeyID])

		case types.EntityDataRepository:
			item, err = qi.api.DataRepository().GetByID(keys[types.KeyID])

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
