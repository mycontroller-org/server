package model

import "time"

// Node model
type Node struct {
	ID        string                 `json:"id"`
	ShortID   string                 `json:"shortId"`
	GatewayID string                 `json:"gatewayId"`
	Name      string                 `json:"name"`
	ParentID  string                 `json:"parentId"`
	LastSeen  time.Time              `json:"lastSeen"`
	Config    map[string]interface{} `json:"config"`
	Others    map[string]interface{} `json:"others"`
	Labels    map[string]string      `json:"labels"`
}

// Sensor model
type Sensor struct {
	ID        string                 `json:"id"`
	ShortID   string                 `json:"shortId"`
	GatewayID string                 `json:"gatewayId"`
	NodeID    string                 `json:"nodeId"`
	Name      string                 `json:"name"`
	LastSeen  time.Time              `json:"lastSeen"`
	Config    map[string]interface{} `json:"config"`
	Others    map[string]interface{} `json:"others"`
}

// SensorField model
type SensorField struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	ShortID         string                 `json:"shortId"`
	GatewayID       string                 `json:"gatewayId"`
	NodeID          string                 `json:"nodeId"`
	SensorID        string                 `json:"sensorId"`
	IsReadOnly      bool                   `json:"isReadOnly"`
	PayloadType     string                 `json:"payloadType"`
	UnitID          string                 `json:"unitId"`
	Payload         FieldValue             `json:"payload"`
	PreviousPayload FieldValue             `json:"previousPayload"`
	LastSeen        time.Time              `json:"lastSeen"`
	Config          map[string]interface{} `json:"config"`
	Others          map[string]interface{} `json:"others"`
	Labels          map[string]string      `json:"labels"`
}

// FieldValue model
type FieldValue struct {
	Value      interface{} `json:"value"`
	IsReceived bool        `json:"isReceived"`
	Timestamp  time.Time   `json:"timestamp"`
}

// config map keys
const (
	CfgUpdateName = "updateName"
)
