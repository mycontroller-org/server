package field

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Field model
type Field struct {
	ID              string               `json:"id"`
	GatewayID       string               `json:"gatewayId"`
	NodeID          string               `json:"nodeId"`
	SensorID        string               `json:"sensorId"`
	FieldID         string               `json:"fieldId"`
	Name            string               `json:"name"`
	MetricType      string               `json:"type"`
	Payload         Payload              `json:"payload"`
	PreviousPayload Payload              `json:"previousPayload"`
	Unit            string               `json:"unit"`
	Labels          cmap.CustomStringMap `json:"labels"`
	Others          cmap.CustomMap       `json:"others"`
	LastSeen        time.Time            `json:"lastSeen"`
}

// Payload model
type Payload struct {
	Value      interface{} `json:"value"`
	IsReceived bool        `json:"isReceived"`
	Timestamp  time.Time   `json:"timestamp"`
}
