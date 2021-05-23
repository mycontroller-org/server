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

	switch payload.Name {
	case nodeML.ActionReboot:
		// TODO: check espnode api to perform a reboot operation
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
	}

	if actionRequest != nil {
		return espNode.Post(actionRequest)
	}
	return nil
}
