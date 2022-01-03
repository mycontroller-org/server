package field

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Field struct
type Field struct {
	ID            string               `json:"id"`
	GatewayID     string               `json:"gatewayId"`
	NodeID        string               `json:"nodeId"`
	SourceID      string               `json:"sourceId"`
	FieldID       string               `json:"fieldId"`
	Name          string               `json:"name"`
	MetricType    string               `json:"metricType"`
	Current       Payload              `json:"current"`
	Previous      Payload              `json:"previous"`
	Formatter     PayloadFormatter     `json:"formatter"`
	Unit          string               `json:"unit"`
	Labels        cmap.CustomStringMap `json:"labels"`
	Others        cmap.CustomMap       `json:"others"`
	NoChangeSince time.Time            `json:"noChangeSince"`
	LastSeen      time.Time            `json:"lastSeen"`
	ModifiedOn    time.Time            `json:"modifiedOn"`
}

// Payload struct
type Payload struct {
	Value      interface{} `json:"value"`
	IsReceived bool        `json:"isReceived"`
	Timestamp  time.Time   `json:"timestamp"`
}

// PayloadFormatter struct
type PayloadFormatter struct {
	OnReceive string `json:"onReceive"`
}
