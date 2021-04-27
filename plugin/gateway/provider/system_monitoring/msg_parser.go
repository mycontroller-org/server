package systemmonitoring

import (
	"errors"

	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
)

// ToRawMessage func implementation
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	if len(msg.Payloads) == 0 {
		return nil, errors.New("there is no payload details on the message")
	}

	if p.NodeID != msg.NodeID {
		return nil, nil
	}

	payload := msg.Payloads[0]

	if msg.Type == msgml.TypeAction {
		switch payload.Name {
		case nodeML.ActionRefreshNodeInfo:
			p.updateNodeDetails()

		}
	}
	return nil, nil
}

// ToMessage implementation
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	return nil, nil
}
