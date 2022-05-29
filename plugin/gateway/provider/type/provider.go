package provider

import (
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
)

// Plugin interface
type Plugin interface {
	Name() string
	Start(rawMessageReceiveFunc func(rawMsg *msgTY.RawMessage) error) error   // start the provider
	Close() error                                                             // close the provider connection
	Post(message *msgTY.Message) error                                        // post a message to the provider
	ConvertToMessages(rawMessage *msgTY.RawMessage) ([]*msgTY.Message, error) // convert the raw message in to Message(s) format
}
