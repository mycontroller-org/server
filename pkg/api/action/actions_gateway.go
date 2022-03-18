package action

import (
	"fmt"

	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

func toGateway(id, action string) error {
	api := resourceAPI{
		Enable:  gatewayAPI.Enable,
		Disable: gatewayAPI.Disable,
		Reload:  gatewayAPI.Reload,
	}
	return toResource(api, id, action)
}

// ExecuteGatewayAction for gateways
func ExecuteGatewayAction(action string, nodeIDs []string) error {
	// verify is a valid action?
	switch action {
	case gatewayTY.ActionDiscoverNodes:
		// nothing to do, just continue
	default:
		return fmt.Errorf("invalid gateway action:%s", action)
	}

	gateways, err := gatewayAPI.GetByIDs(nodeIDs)
	if err != nil {
		return err
	}
	for index := 0; index < len(gateways); index++ {
		gateway := gateways[index]
		if !gateway.Enabled {
			continue
		}
		err = toGatewayAction(gateway.ID, action)
		if err != nil {
			zap.L().Error("error on sending an action to a gateway", zap.Error(err), zap.String("gateway", gateway.ID))
		}
	}
	return nil
}

func toGatewayAction(gatewayID, action string) error {
	msg := msgTY.NewMessage(false)
	msg.GatewayID = gatewayID
	pl := msgTY.NewPayload()
	pl.Key = action
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgTY.TypeAction
	return Post(&msg)
}
