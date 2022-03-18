package esphome

import (
	"time"

	esphomeAPI "github.com/mycontroller-org/esphome_api/pkg/api"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// handleActions performs actions on a node
func (p *Provider) handleActions(message *msgTY.Message) error {
	pl := message.Payloads[0]
	var actionRequest protoreflect.ProtoMessage
	espNode := p.clientStore.Get(message.NodeID)
	if espNode == nil {
		return nil
	}

	switch pl.Key {
	case nodeTY.ActionReboot:
		// I do not see a direct option to reboot or restart
		// a entity should be created on the espnose as mentioned in https://esphome.io/components/switch/restart.html
		// Note: name of the switch should as 'restart'
		espNode.sendRestartRequest()
		return nil

	case nodeTY.ActionRefreshNodeInfo:
		espNode.sendNodeInfo()
		return nil

	case nodeTY.ActionHeartbeatRequest:
		espNode.doAliveCheck()
		return nil

	case ActionTimeRequest:
		actionRequest = &esphomeAPI.GetTimeResponse{
			EpochSeconds: uint32(time.Now().Unix()),
		}

	case ActionPingRequest:
		actionRequest = &esphomeAPI.PingResponse{}

	default:
		// noop
	}

	if actionRequest != nil {
		return espNode.Post(actionRequest)
	}
	return nil
}
