package message

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Message definition
type Message struct {
	ID            string
	GatewayID     string
	NodeID        string
	SensorID      string
	Type          string // Message type: set, request, ...
	FieldName     string // name of the field only for field data, for node, senor can be any thing
	MetricType    string // none, binary, gauge, counter, geo ...
	Payload       string // 1, true, 99.45, started, 72.345,45.333, any...
	Unit          string // volt, milli_volt, etc...
	IsAck         bool   // Is this acknowledgement message
	IsReceived    bool   // Is this received message
	IsAckEnabled  bool   // Is Acknowledge enabled?
	IsPassiveNode bool   // Is this message for passive node or sleeping node?
	Timestamp     time.Time
	Labels        cmap.CustomStringMap
	Others        cmap.CustomMap
}

// Clone a message
func (m *Message) Clone() *Message {
	cm := &Message{
		ID:            m.ID,
		GatewayID:     m.GatewayID,
		NodeID:        m.NodeID,
		SensorID:      m.SensorID,
		Type:          m.Type,
		FieldName:     m.FieldName,
		MetricType:    m.MetricType,
		Payload:       m.Payload,
		Unit:          m.Unit,
		IsAck:         m.IsAck,
		IsReceived:    m.IsReceived,
		IsAckEnabled:  m.IsAckEnabled,
		IsPassiveNode: m.IsPassiveNode,
		Timestamp:     m.Timestamp,
		Labels:        m.Labels.Clone(),
		Others:        m.Others.Clone(),
	}
	return cm
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
	ID         string
	IsReceived bool
	Data       []byte
	Timestamp  time.Time
	Others     cmap.CustomMap
}

// MarshalJSON for RawMessage, Data should be printed as string on the log
func (rm *RawMessage) MarshalJSON() ([]byte, error) {
	type RawMsgAlias RawMessage
	return json.Marshal(&struct {
		Data string
		*RawMsgAlias
	}{
		Data:        string(rm.Data),
		RawMsgAlias: (*RawMsgAlias)(rm),
	})
}

// Clone func
func (rm *RawMessage) Clone() *RawMessage {
	return &RawMessage{
		ID:         rm.ID,
		IsReceived: rm.IsReceived,
		Data:       rm.Data,
		Timestamp:  rm.Timestamp,
		Others:     rm.Others.Clone(),
	}
}

// DeliveryStatus definition
type DeliveryStatus struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}