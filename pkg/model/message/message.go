package message

import (
	"bytes"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
)

// Message definition
type Message struct {
	ID            string
	GatewayID     string
	NodeID        string
	SensorID      string
	Type          string // Message type: set, request, ...
	FieldName     string // name of the field, only for field data
	MetricType    string // none, binary, gauge, counter, geo ...
	Payload       string // 1, true, 99.45, started, 72.345,45.333, any...
	Unit          string // volt, milli_volt, etc...
	IsAck         bool   // Is this acknowledgement message
	IsReceived    bool   // Is this received message
	IsAckEnabled  bool   // Is Acknowledge enabled?
	IsPassiveNode bool   // Is this message for passive node or sleeping node?
	Timestamp     time.Time
	Labels        ml.CustomStringMap
	Others        ml.CustomMap
}

//GetID returns unique ID for this message
func (m *Message) GetID() string {
	var buffer bytes.Buffer
	buffer.WriteString(m.GatewayID)

	if m.NodeID != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.NodeID)
	}
	if m.SensorID != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.SensorID)
	}
	if m.FieldName != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.FieldName)
	}
	if m.Type != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.Type)
	}
	return buffer.String()
}

// RawMessage from/to gateway media
type RawMessage struct {
	ID        string
	Data      []byte
	Timestamp time.Time
	Others    ml.CustomMap
}

// DeliveryStatus definition
type DeliveryStatus struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}
