package provider

import (
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

// Provider interface
type Provider interface {
	ToRawMessage(message *msgml.Message) (*msgml.RawMessage, error)
	ToMessage(rawMesage *msgml.RawMessage) ([]*msgml.Message, error)
	Post(rawMessage *msgml.RawMessage) error
	Start(messageReceiveFunc func(rawMsg *msgml.RawMessage) error) error
	Close() error
}

// Providers list
const (
	TypeMySensors = "mysensors"
	TypeTasmota   = "tasmota"
)
