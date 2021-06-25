package action

import (
	handlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
)

func toHandler(id, action string) error {
	api := resourceAPI{
		Enable:  handlerAPI.Enable,
		Disable: handlerAPI.Disable,
		Reload:  handlerAPI.Reload,
	}
	return toResource(api, id, action)
}
