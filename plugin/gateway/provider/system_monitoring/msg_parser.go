package systemmonitoring

import (
	"errors"

	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
)

// ProcessReceived implementation
func (p *Provider) ProcessReceived(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	// gateway do not send a messages to queue, sends directly
	return nil, nil
}

// Post func
func (p *Provider) Post(msg *msgml.Message) error {
	if len(msg.Payloads) == 0 {
		return errors.New("there is no payload details on the message")
	}

	if p.NodeID != msg.NodeID {
		return nil
	}

	payload := msg.Payloads[0]

	if msg.Type == msgml.TypeAction {
		switch payload.Key {
		case nodeML.ActionRefreshNodeInfo:
			p.updateNodeDetails()

		}
	}
	return nil
}
