package message

import (
	"bytes"
	"time"

	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Data definition
type Data struct {
	Name       string               // name of the field only for field data, for node, senor can be any thing
	Value      string               // 1, true, 99.45, started, 72.345,45.333, any...
	MetricType string               // none, binary, gauge, counter, geo ...
	Unit       string               // volt, milli_volt, etc...
	Labels     cmap.CustomStringMap // labels for this field or data
	Others     cmap.CustomMap       // can be used to store other details
}

// NewData returns empty data
func NewData() Data {
	data := Data{}
	data.Labels = data.Labels.Init()
	data.Others = data.Others.Init()
	return data
}

// Clone a data
func (d *Data) Clone() Data {
	cd := Data{
		Name:       d.Name,
		Value:      d.Value,
		MetricType: d.MetricType,
		Unit:       d.Unit,
		Labels:     d.Labels.Clone(),
		Others:     d.Others.Clone(),
	}
	return cd
}

// Message definition
type Message struct {
	ID            string
	GatewayID     string
	NodeID        string
	SourceID      string
	Type          string // Message type: set, request, ...
	Payloads      []Data // payloads
	IsAck         bool   // Is this acknowledgement message
	IsReceived    bool   // Is this received message
	IsAckEnabled  bool   // Is Acknowledge enabled?
	IsPassiveNode bool   // Is this message for passive node or sleeping node?
	Timestamp     time.Time
}

// NewMessage returns empty message
func NewMessage(isReceived bool) Message {
	return Message{IsReceived: isReceived, Payloads: make([]Data, 0)}
}

// Clone a message
func (m *Message) Clone() *Message {
	// clone data slice
	clonedData := make([]Data, 0)
	for _, d := range m.Payloads {
		clonedData = append(clonedData, d.Clone())
	}
	cm := &Message{
		ID:            m.ID,
		GatewayID:     m.GatewayID,
		NodeID:        m.NodeID,
		SourceID:      m.SourceID,
		Type:          m.Type,
		Payloads:      clonedData,
		IsAck:         m.IsAck,
		IsReceived:    m.IsReceived,
		IsAckEnabled:  m.IsAckEnabled,
		IsPassiveNode: m.IsPassiveNode,
		Timestamp:     m.Timestamp,
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
	if m.SourceID != "" {
		buffer.WriteString("-")
		buffer.WriteString(m.SourceID)
	}
	if len(m.Payloads) > 0 {
		for _, d := range m.Payloads {
			if d.Name != "" {
				buffer.WriteString("-")
				buffer.WriteString(d.Name)
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
	ID                 string
	IsReceived         bool
	AcknowledgeEnabled bool
	Data               []byte
	Timestamp          time.Time
	Others             cmap.CustomMap
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
