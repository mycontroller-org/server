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
	IsAck      bool // Is this acknowledgement message
	IsReceived bool // Is this received message
	Command    string
	SubCommand string
	Field      string
	Payload    string
	DataType   string
	UnitID     string
	Timestamp  time.Time
	Others     map[string]interface{}
}

// RawMessage from/to gateway media
type RawMessage struct {
	ID        string
	Data      []byte
	Timestamp time.Time
	Others    map[string]interface{}
}

// DeliveryStatus definition
type DeliveryStatus struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}
