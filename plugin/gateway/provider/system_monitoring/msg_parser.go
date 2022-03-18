package systemmonitoring

import (
	"errors"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
)

// ProcessReceived implementation
func (p *Provider) ProcessReceived(rawMsg *msgTY.RawMessage) ([]*msgTY.Message, error) {
	// gateway do not send a messages to queue, sends directly
	return nil, nil
}

// Post func
func (p *Provider) Post(msg *msgTY.Message) error {
	if len(msg.Payloads) == 0 {
		return errors.New("there is no payload details on the message")
	}

	if p.NodeID != msg.NodeID {
		return nil
	}

	payload := msg.Payloads[0]

	if msg.Type == msgTY.TypeAction {
		switch payload.Key {
		case nodeTY.ActionRefreshNodeInfo:
			p.updateNodeDetails()

		}
	}
	return nil
}
