package br_map

import (
	"fmt"

	dashboardAPI "github.com/mycontroller-org/server/v2/pkg/api/dashboard"
	dataRepositoryAPI "github.com/mycontroller-org/server/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	fwAPI "github.com/mycontroller-org/server/v2/pkg/api/firmware"
	fwdPayloadAPI "github.com/mycontroller-org/server/v2/pkg/api/forward_payload"
	gwAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	handlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
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
	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	dataRepositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
)

// import map
var (
	ImportMap = map[string]backupTY.SaveAPIHolder{
		types.EntityDashboard: {
			EntityType: dashboardTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(dashboardTY.Config); ok {
					return dashboardAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityDataRepository: {
			EntityType: dataRepositoryTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(dataRepositoryTY.Config); ok {
					return dataRepositoryAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityField: {
			EntityType: fieldTY.Field{},
			API: func(data interface{}) error {
				if input, ok := data.(fieldTY.Field); ok {
					return fieldAPI.Save(&input, false)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityFirmware: {
			EntityType: firmwareTY.Firmware{},
			API: func(data interface{}) error {
				if input, ok := data.(firmwareTY.Firmware); ok {
					return fwAPI.Save(&input, false)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityForwardPayload: {
			EntityType: fwdPayloadTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(fwdPayloadTY.Config); ok {
					return fwdPayloadAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityGateway: {
			EntityType: gatewayTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(gatewayTY.Config); ok {
					return gwAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityHandler: {
			EntityType: handlerTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(handlerTY.Config); ok {
					return handlerAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityNode: {
			EntityType: nodeTY.Node{},
			API: func(data interface{}) error {
				if input, ok := data.(nodeTY.Node); ok {
					return nodeAPI.Save(&input, false)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntitySchedule: {
			EntityType: scheduleTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(scheduleTY.Config); ok {
					return scheduleAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntitySettings: {
			EntityType: settingsTY.Settings{},
			API: func(data interface{}) error {
				if input, ok := data.(settingsTY.Settings); ok {
					return settingsAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntitySource: {
			EntityType: sourceTY.Source{},
			API: func(data interface{}) error {
				if input, ok := data.(sourceTY.Source); ok {
					return sourceAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityTask: {
			EntityType: taskTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(taskTY.Config); ok {
					return taskAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityUser: {
			EntityType: userTY.User{},
			API: func(data interface{}) error {
				if input, ok := data.(userTY.User); ok {
					return userAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityVirtualAssistant: {
			EntityType: vaTY.Config{},
			API: func(data interface{}) error {
				if input, ok := data.(vaTY.Config); ok {
					return vaAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityVirtualDevice: {
			EntityType: vdTY.VirtualDevice{},
			API: func(data interface{}) error {
				if input, ok := data.(vdTY.VirtualDevice); ok {
					return vdAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},

		types.EntityServiceToken: {
			EntityType: svcTokenTY.ServiceToken{},
			API: func(data interface{}) error {
				if input, ok := data.(svcTokenTY.ServiceToken); ok {
					return svcTokenAPI.Save(&input)
				}
				return fmt.Errorf("invalid type:%T", data)
			},
		},
	}
)
