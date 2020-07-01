package message

import "time"

// Wrapper for queue
type Wrapper struct {
	GatewayID  string
	IsReceived bool
	Message    interface{}
}

// Message definition
type Message struct {
	ID         string
	GatewayID  string
	NodeID     string
	SensorID   string
	Fields     []Field
	IsAck      bool // Is this acknowledgement message
	IsReceived bool // Is this received message
	Timestamp  time.Time
	Others     map[string]interface{}
}

// RawMessage from/to gateway media
type RawMessage struct {
	Data      []byte
	Timestamp time.Time
	Others    map[string]interface{}
}

// Field definition
type Field struct {
	Key      string
	Payload  string
	Command  string
	DataType string
	UnitID   string
	Others   map[string]interface{}
}

// DeliveryStatus definition
type DeliveryStatus struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}
