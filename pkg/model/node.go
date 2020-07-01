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
	Others    map[string]interface{} `json:"others"`
	Config    map[string]interface{} `json:"config"`
}

// Sensor model
type Sensor struct {
	ID             string                 `json:"id"`
	ShortID        string                 `json:"shortId"`
	GatewayID      string                 `json:"gatewayId"`
	NodeID         string                 `json:"nodeId"`
	Name           string                 `json:"name"`
	LastSeen       time.Time              `json:"lastSeen"`
	Others         map[string]interface{} `json:"others"`
	Config         map[string]interface{} `json:"config"`
	ProviderConfig map[string]interface{} `json:"providerConfig"`
}

// SensorField model
type SensorField struct {
	ID              string                 `json:"id"`
	ShortID         string                 `json:"shortId"`
	GatewayID       string                 `json:"gatewayId"`
	NodeID          string                 `json:"nodeId"`
	SensorID        string                 `json:"sensorId"`
	IsReadOnly      bool                   `json:"isReadOnly"`
	DataType        string                 `json:"dataType"`
	UnitID          string                 `json:"unitId"`
	Payload         FieldValue             `json:"payload"`
	PreviousPayload FieldValue             `json:"previousPayload"`
	LastSeen        time.Time              `json:"lastSeen"`
	Others          map[string]interface{} `json:"others"`
	Config          map[string]interface{} `json:"config"`
	ProviderConfig  map[string]interface{} `json:"providerConfig"`
}

// FieldValue model
type FieldValue struct {
	Value      interface{} `json:"value"`
	IsReceived bool        `json:"isReceived"`
	Timestamp  time.Time   `json:"timestamp"`
}
