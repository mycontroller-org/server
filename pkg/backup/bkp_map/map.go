package br_map

import (
	"context"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
)

// returns import and export entity map
func GetStorageApiMap(ctx context.Context) (map[string]backupTY.Backup, error) {
	entities, err := entitiesAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	funcMap := map[string]backupTY.Backup{
		types.EntityDashboard:        entities.Dashboard(),
		types.EntityDataRepository:   entities.DataRepository(),
		types.EntityField:            entities.Field(),
		types.EntityFirmware:         entities.Firmware(),
		types.EntityForwardPayload:   entities.ForwardPayload(),
		types.EntityGateway:          entities.Gateway(),
		types.EntityHandler:          entities.Handler(),
		types.EntityNode:             entities.Node(),
		types.EntitySchedule:         entities.Schedule(),
		types.EntitySettings:         entities.Settings(),
		types.EntitySource:           entities.Source(),
		types.EntityTask:             entities.Task(),
		types.EntityUser:             entities.User(),
		types.EntityVirtualAssistant: entities.VirtualAssistant(),
		types.EntityVirtualDevice:    entities.VirtualDevice(),
		types.EntityServiceToken:     entities.ServiceToken(),
	}

	return funcMap, nil
}
