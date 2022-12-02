package field

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Field struct
type Field struct {
	ID            string               `json:"id" yaml:"id"`
	GatewayID     string               `json:"gatewayId" yaml:"gatewayId"`
	NodeID        string               `json:"nodeId" yaml:"nodeId"`
	SourceID      string               `json:"sourceId" yaml:"sourceId"`
	FieldID       string               `json:"fieldId" yaml:"fieldId"`
	Name          string               `json:"name" yaml:"name"`
	MetricType    string               `json:"metricType" yaml:"metricType"`
	Current       Payload              `json:"current" yaml:"current"`
	Previous      Payload              `json:"previous" yaml:"previous"`
	Formatter     PayloadFormatter     `json:"formatter" yaml:"formatter"`
	Unit          string               `json:"unit" yaml:"unit"`
	Labels        cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Others        cmap.CustomMap       `json:"others" yaml:"others"`
	NoChangeSince time.Time            `json:"noChangeSince" yaml:"noChangeSince"`
	LastSeen      time.Time            `json:"lastSeen" yaml:"lastSeen"`
	ModifiedOn    time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}

// Payload struct
type Payload struct {
	Value      interface{} `json:"value" yaml:"value"`
	IsReceived bool        `json:"isReceived" yaml:"isReceived"`
	Timestamp  time.Time   `json:"timestamp" yaml:"timestamp"`
}

// PayloadFormatter struct
type PayloadFormatter struct {
	OnReceive string `json:"onReceive" yaml:"onReceive"`
}

// clones field
func (f *Field) Clone() *Field {
	return &Field{
		ID:         f.ID,
		GatewayID:  f.GatewayID,
		NodeID:     f.NodeID,
		SourceID:   f.SourceID,
		FieldID:    f.FieldID,
		Name:       f.Name,
		MetricType: f.MetricType,
		Unit:       f.Unit,
		Current:    f.Current,
		Previous:   f.Previous,
		Labels:     f.Labels.Clone(),
		Others:     f.Others.Clone(),
	}
}
