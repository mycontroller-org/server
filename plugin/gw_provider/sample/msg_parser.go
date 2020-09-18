package sample

import msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"

// implement message parser

// ToRawMessage func implementation
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	return nil, nil
}

// ToMessage implementation
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	return nil, nil
}
