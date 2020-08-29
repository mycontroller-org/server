package gatewayprotocol

import msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"

// Global constants
const (
	// Gateway Types
	TypeMQTT     = "mqtt"
	TypeSerial   = "serial"
	TypeEthernet = "ethernet"
)

// Others map known keys
const (
	KeyTopic = "topic"
	KeyQoS   = "qos"
	KeyName  = "name"
)

// Gateway protocol interface
type Gateway interface {
	Write(rawMsg *msgml.RawMessage) error
	Close() error
}
