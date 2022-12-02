package message

import (
	"bytes"
	"time"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	customTY "github.com/mycontroller-org/server/v2/pkg/types/custom_types"
	cloneutil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
)

// Payload definition
type Payload struct {
	Key        string               `json:"key" yaml:"key"`               // it is id for the the fields. for node, source key says what it is
	Value      customTY.StringData  `json:"value" yaml:"value"`           // 1, true, 99.45, started, 72.345,45.333, any...
	MetricType string               `json:"metricType" yaml:"metricType"` // none, binary, gauge, counter, geo ...
	Unit       string               `json:"unit" yaml:"unit"`             // Volt, MilliVolt, °C, °F, ⚡, etc...
	Labels     cmap.CustomStringMap `json:"labels" yaml:"labels"`         // labels for this data
	Others     cmap.CustomMap       `json:"others" yaml:"others"`         // can hold other than key, value data.
}

// NewPayload returns empty payload
func NewPayload() Payload {
	data := Payload{}
	data.Labels = data.Labels.Init()
	data.Others = data.Others.Init()
	return data
}

// sets payload value
func (p *Payload) SetValue(value string) {
	p.Value = customTY.StringData(value)
}

// Clone a data
func (p *Payload) Clone() Payload {
	cd := Payload{
		Key:        p.Key,
		Value:      p.Value,
		MetricType: p.MetricType,
		Unit:       p.Unit,
		Labels:     p.Labels.Clone(),
		Others:     p.Others.Clone(),
	}
	return cd
}

// Message definition
type Message struct {
	ID           string               `json:"id" yaml:"id"`
	GatewayID    string               `json:"gatewayId" yaml:"gatewayId"`
	NodeID       string               `json:"nodeId" yaml:"nodeId"`
	SourceID     string               `json:"sourceId" yaml:"sourceId"`
	Type         string               `json:"type" yaml:"type"`                 // Message type: set, request, ...
	Payloads     []Payload            `json:"payloads" yaml:"payloads"`         // payloads
	Labels       cmap.CustomStringMap `json:"labels" yaml:"labels"`             // these labels can be used in transport or other purpose
	IsAck        bool                 `json:"isAck" yaml:"isAck"`               // Is this acknowledgement message
	IsReceived   bool                 `json:"isReceived" yaml:"isReceived"`     // Is this received message
	IsAckEnabled bool                 `json:"isAckEnabled" yaml:"isAckEnabled"` // Is Acknowledge enabled?
	IsSleepNode  bool                 `json:"isSleepNode" yaml:"isSleepNode"`   // Is this message for active node or sleep node?
	Timestamp    time.Time            `json:"timestamp" yaml:"timestamp"`
}

// NewMessage returns empty message
func NewMessage(isReceived bool) Message {
	msg := Message{IsReceived: isReceived, Payloads: make([]Payload, 0)}
	msg.Labels = msg.Labels.Init()
	return msg
}

// Clone a message
func (m *Message) Clone() *Message {
	// clone data slice
	clonedData := make([]Payload, 0)
	for _, d := range m.Payloads {
		clonedData = append(clonedData, d.Clone())
	}
	m.Labels = m.Labels.Init() // init labels if not done
	cm := &Message{
		ID:           m.ID,
		GatewayID:    m.GatewayID,
		NodeID:       m.NodeID,
		SourceID:     m.SourceID,
		Type:         m.Type,
		Payloads:     clonedData,
		Labels:       m.Labels.Clone(),
		IsAck:        m.IsAck,
		IsReceived:   m.IsReceived,
		IsAckEnabled: m.IsAckEnabled,
		IsSleepNode:  m.IsSleepNode,
		Timestamp:    m.Timestamp,
	}
	return cm
}

// GetID returns unique ID for this message
func (m *Message) GetID() string {
	var buffer bytes.Buffer
	buffer.WriteString(m.GatewayID)

	if m.NodeID != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.NodeID)
	}
	if m.SourceID != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.SourceID)
	}
	if len(m.Payloads) > 0 {
		for _, d := range m.Payloads {
			if d.Key != "" {
				buffer.WriteString("-")
				buffer.WriteString(d.Key)
			}
		}
	}

	if m.Type != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.Type)
	}
	return buffer.String()
}

// RawMessage from/to gateway media
type RawMessage struct {
	ID           string         `json:"id" yaml:"id"`
	IsReceived   bool           `json:"isReceived" yaml:"isReceived"`
	IsAckEnabled bool           `json:"isAckEnabled" yaml:"isAckEnabled"`
	Data         interface{}    `json:"data" yaml:"data"`
	Timestamp    time.Time      `json:"timestamp" yaml:"timestamp"`
	Others       cmap.CustomMap `json:"others" yaml:"others"`
}

// NewRawMessage returns empty message
func NewRawMessage(isReceived bool, data []byte) *RawMessage {
	rawMsg := &RawMessage{
		Timestamp:  time.Now(),
		IsReceived: isReceived,
		Data:       data,
	}
	rawMsg.Others = rawMsg.Others.Init()
	return rawMsg
}

// MarshalJSON for RawMessage, Data should be printed as string on the log
func (rm *RawMessage) MarshalJSON() ([]byte, error) {
	type RawMsgAlias RawMessage
	return json.Marshal(&struct {
		Data string
		*RawMsgAlias
	}{
		Data:        convertor.ToString(rm.Data),
		RawMsgAlias: (*RawMsgAlias)(rm),
	})
}

// Clone func
func (rm *RawMessage) Clone() *RawMessage {
	return &RawMessage{
		ID:         rm.ID,
		IsReceived: rm.IsReceived,
		Data:       cloneutil.Clone(rm.Data),
		Timestamp:  rm.Timestamp,
		Others:     rm.Others.Clone(),
	}
}

// DeliveryStatus definition
type DeliveryStatus struct {
	ID      string `json:"id" yaml:"id"`
	Success bool   `json:"success" yaml:"success"`
	Message string `json:"message" yaml:"message"`
}
