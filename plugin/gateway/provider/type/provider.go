package provider

import (
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
)

// Plugin interface
type Plugin interface {
	Name() string
	Start(messageReceiveFunc func(rawMsg *msgML.RawMessage) error) error   // start the provider
	Close() error                                                          // close the provider connection
	Post(message *msgML.Message) error                                     // post a message to the provider
	ProcessReceived(rawMesage *msgML.RawMessage) ([]*msgML.Message, error) // process the received messages
}
