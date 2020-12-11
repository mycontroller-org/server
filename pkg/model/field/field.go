package field

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Field model
type Field struct {
	ID               string               `json:"id"`
	GatewayID        string               `json:"gatewayId"`
	NodeID           string               `json:"nodeId"`
	SensorID         string               `json:"sensorId"`
	FieldID          string               `json:"fieldId"`
	Name             string               `json:"name"`
	MetricType       string               `json:"metricType"`
	Payload          Payload              `json:"payload"`
	PreviousPayload  Payload              `json:"previousPayload"`
	Unit             string               `json:"unit"`
	Labels           cmap.CustomStringMap `json:"labels"`
	Others           cmap.CustomMap       `json:"others"`
	NoChangeSince    time.Time            `json:"noChangeSince"`
	PayloadFormatter PayloadFormatter     `json:"payloadFormatter"`
	LastSeen         time.Time            `json:"lastSeen"`
	LastModifiedOn   time.Time            `json:"lastModifiedOn"`
}

// Payload model
type Payload struct {
	Value      interface{} `json:"value"`
	IsReceived bool        `json:"isReceived"`
	Timestamp  time.Time   `json:"timestamp"`
}

// PayloadFormatter model
type PayloadFormatter struct {
	OnReceive string `json:"onReceive"`
}
