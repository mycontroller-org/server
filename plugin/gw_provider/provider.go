package provider

import (
	"fmt"

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

// Topics used across provide componenet
const (
	TopicMessagePostToCore     = "message_to_core"   // posts message in to core component
	TopicMessageListenFromCore = "message_from_core" // receives messages from core component
)

// GetTopicListenFromProcessor returns listen topic
func GetTopicListenFromProcessor(gatewayID string) string {
	return fmt.Sprintf("%s_%s", TopicMessageListenFromCore, gatewayID)
}
