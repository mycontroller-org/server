package protocol

import (
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
)

// Protocol interface
type Protocol interface {
	Write(rawMsg *msgTY.RawMessage) error // write a message on a specified protocol
	Close() error                         // close the protocol connection
}

// Protocol Types
const (
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
