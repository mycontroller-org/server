package field

import (
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
)

// Field model
type Field struct {
	ID              string             `json:"id"`
	GatewayID       string             `json:"gatewayId"`
	NodeID          string             `json:"nodeId"`
	SensorID        string             `json:"sensorId"`
	Name            string             `json:"name"`
	MetricType      string             `json:"type"`
	Payload         Payload            `json:"payload"`
	PreviousPayload Payload            `json:"previousPayload"`
	Unit            string             `json:"unit"`
	Labels          ml.CustomStringMap `json:"labels"`
	Others          ml.CustomMap       `json:"others"`
	LastSeen        time.Time          `json:"lastSeen"`
}

// Payload model
type Payload struct {
	Value      interface{} `json:"value"`
	IsReceived bool        `json:"isReceived"`
	Timestamp  time.Time   `json:"timestamp"`
}
