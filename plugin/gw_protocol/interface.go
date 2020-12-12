package gwprotocol

import msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"

// Gateway protocol interface
type Gateway interface {
	Write(rawMsg *msgml.RawMessage) error // write a message to gateway
	Close() error                         // close the gateway connection
}

// Global constants
const (
	// Gateway Types
	TypeMQTT     = "mqtt"
	TypeSerial   = "serial"
	TypeEthernet = "ethernet"
)

// Others map known keys
const (
	// mqtt requirements
	KeyMqttTopic = "mqtt_topic"
	KeyMqttQoS   = "mqtt_qos"

	// http requirements
	KeyHTTPRequestConf  = "http_request_conf"
	KeyHTTPResponseConf = "http_response_conf"
)
