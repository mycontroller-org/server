package esphome

import (
	"time"

	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	esphomeAPI "github.com/mycontroller-org/esphome_api/pkg/api"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// handleActions performs actions on a node
func (p *Provider) handleActions(message *msgML.Message) error {
	payload := message.Payloads[0]
	var actionRequest protoreflect.ProtoMessage
	espNode := p.clientStore.Get(message.NodeID)
	if espNode == nil {
		return nil
	}

	switch payload.Key {
	case nodeML.ActionReboot:
		// I do not see a direct option to reboot or restart
		// a entity should be created on the espnose as mentioned in https://esphome.io/components/switch/restart.html
		// Note: name of the switch should as 'restart'
		espNode.sendRestartRequest()
		return nil

	case nodeML.ActionRefreshNodeInfo:
		espNode.sendNodeInfo()
		return nil

	case nodeML.ActionHeartbeatRequest:
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
