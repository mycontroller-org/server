package action

import (
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
)

func toGateway(id, action string) error {
	api := resourceAPI{
		Enable:  gatewayAPI.Enable,
		Disable: gatewayAPI.Disable,
		Reload:  gatewayAPI.Reload,
	}
	return toResource(api, id, action)
}
