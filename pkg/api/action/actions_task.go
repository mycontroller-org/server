package action

import (
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
)

func toTask(id, action string) error {
	api := resourceAPI{
		Enable:  taskAPI.Enable,
		Disable: taskAPI.Disable,
		Reload:  taskAPI.Reload,
	}
	return toResource(api, id, action)
}
