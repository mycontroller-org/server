package action

import (
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
)

func toSchedule(id, action string) error {
	api := resourceAPI{
		Enable:  schedulerAPI.Enable,
		Disable: schedulerAPI.Disable,
		Reload:  schedulerAPI.Reload,
	}
	return toResource(api, id, action)
}
