package action

import (
	scheduleAPI "github.com/mycontroller-org/backend/v2/pkg/api/schedule"
)

func toSchedule(id, action string) error {
	api := resourceAPI{
		Enable:  scheduleAPI.Enable,
		Disable: scheduleAPI.Disable,
		Reload:  scheduleAPI.Reload,
	}
	return toResource(api, id, action)
}
