package br_map

import (
	dashboardAPI "github.com/mycontroller-org/server/v2/pkg/api/dashboard"
	dataRepositoryAPI "github.com/mycontroller-org/server/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	firmwareAPI "github.com/mycontroller-org/server/v2/pkg/api/firmware"
	forwardPayloadAPI "github.com/mycontroller-org/server/v2/pkg/api/forward_payload"
	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	notificationHandlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	svcTokenAPI "github.com/mycontroller-org/server/v2/pkg/api/service_token"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	sourceAPI "github.com/mycontroller-org/server/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	userAPI "github.com/mycontroller-org/server/v2/pkg/api/user"
	vaAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_assistant"
	vdAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_device"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
)

var (
	ExportMap = map[string]backupTY.ListFunc{
		types.EntityGateway:          gatewayAPI.List,
		types.EntityNode:             nodeAPI.List,
		types.EntitySource:           sourceAPI.List,
		types.EntityField:            fieldAPI.List,
		types.EntityFirmware:         firmwareAPI.List,
		types.EntityUser:             userAPI.List,
		types.EntityDashboard:        dashboardAPI.List,
		types.EntityForwardPayload:   forwardPayloadAPI.List,
		types.EntityHandler:          notificationHandlerAPI.List,
		types.EntityTask:             taskAPI.List,
		types.EntitySchedule:         scheduleAPI.List,
		types.EntitySettings:         settingsAPI.List,
		types.EntityDataRepository:   dataRepositoryAPI.List,
		types.EntityVirtualAssistant: vaAPI.List,
		types.EntityVirtualDevice:    vdAPI.List,
		types.EntityServiceToken:     svcTokenAPI.List,
	}
)
