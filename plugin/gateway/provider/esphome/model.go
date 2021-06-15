package esphome

import (
	"bytes"

	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/backend/v2/plugin/metrics"
	esphomeClient "github.com/mycontroller-org/esphome_api/pkg/client"
)

const (
	// internal references
	MessageTypeID = "message_type_id"
	NodeID        = "node_id"
	LabelKey      = "key"
	LabelType     = "type"

	// camera image codec details
	CameraImagePrefix = "data:image/jpeg;base64,"

	// internal actions
	ActionTimeRequest = "time_request"
	ActionPingRequest = "ping_request"

	// entity types
	EntityTypeBinarySensor = "binary_sensor"
	EntityTypeCamera       = "camera"
	EntityTypeClimate      = "climate"
	EntityTypeCover        = "cover"
	EntityTypeFan          = "fan"
	EntityTypeLight        = "light"
	EntityTypeSensor       = "sensor"
	EntityTypeSwitch       = "switch"
	EntityTypeTextSensor   = "text_sensor"

	// source id for restart
	SourceIDRestart = "restart"

	// response message common fields
	FieldKey      = "key"
	FieldObjectID = "object_id"
	FieldName     = "name"
	FieldUniqueID = "unique_id"

	// response message fields
	FieldState                 = "state"
	FieldBrightness            = "brightness"
	FieldRGB                   = "rgb"
	FieldRed                   = "red"
	FieldGreen                 = "green"
	FieldBlue                  = "blue"
	FieldWhite                 = "white"
	FieldColorTemperature      = "color_temperature"
	FieldEffect                = "effect"
	FieldStream                = "stream"
	FieldData                  = "data"
	FieldMode                  = "mode"
	FieldTargetTemperature     = "target_temperature"
	FieldTargetTemperatureLow  = "target_temperature_low"
	FieldTargetTemperatureHigh = "target_temperature_high"
	FieldAway                  = "away"
	FieldCurrentTemperature    = "current_temperature"
	FieldOscillating           = "oscillating"
	FieldFanMode               = "fan_mode"
	FieldSwingMode             = "swing_mode"
	FieldPosition              = "position"
	FieldTilt                  = "tilt"
	FieldStop                  = "stop"
	FieldSpeed                 = "speed"
	FieldDirection             = "direction"
	FieldSpeedLevel            = "speed_level"
)

// ESPHomeNodeConfig holds esphome node configuration details
type ESPHomeNodeConfig struct {
	Disabled           bool
	Address            string
	Password           string
	Timeout            string
	AliveCheckInterval string
	ReconnectDelay     string
}

// ESPHomeNode is a esphome node instance
type ESPHomeNode struct {
	GatewayID     string
	NodeID        string
	Config        ESPHomeNodeConfig
	Client        *esphomeClient.Client
	rxMessageFunc func(rawMsg *msgML.RawMessage) error
	imageBuffer   *bytes.Buffer
	entityStore   *EntityStore
}

// Entity holds key sourceId details of a entity
type Entity struct {
	Key         uint32
	Type        string
	SourceID    string
	SourceName  string
	DeviceClass string
	Unit        string
}

// Clone produces a cloned version of entity
func (en *Entity) Clone() *Entity {
	return &Entity{
		Key:         en.Key,
		Type:        en.Type,
		SourceID:    en.SourceID,
		SourceName:  en.SourceName,
		DeviceClass: en.DeviceClass,
		Unit:        en.Unit,
	}
}

// MetricMap holds metric details
type MetricMap struct {
	MetricType string
	Unit       string
	ReadOnly   string
}

// newMetricMap creates a MetricMap
func newMetricMap(metricType, unit string, isReadOnly bool) MetricMap {
	return MetricMap{MetricType: metricType, Unit: unit, ReadOnly: convertor.ToString(isReadOnly)}
}

// getMetricData returns metric details for a field based on the entity type
func getMetricData(entityType, fieldID string) MetricMap {
	switch entityType {

	case EntityTypeClimate:
		switch fieldID {
		case FieldCurrentTemperature:
			return newMetricMap(metrics.MetricTypeGaugeFloat, "°C", true)
		case FieldTargetTemperature, FieldTargetTemperatureLow, FieldTargetTemperatureHigh:
			return newMetricMap(metrics.MetricTypeGaugeFloat, "°C", false)
		default:
			return newMetricMap(metrics.MetricTypeNone, "", false)
		}

	case EntityTypeFan:
		switch fieldID {
		case FieldState:
			return newMetricMap(metrics.MetricTypeBinary, "", false)
		default:
			return newMetricMap(metrics.MetricTypeNone, "", false)
		}

	case EntityTypeBinarySensor:
		return newMetricMap(metrics.MetricTypeBinary, "", true)

	case EntityTypeSwitch:
		return newMetricMap(metrics.MetricTypeBinary, "", false)

	case EntityTypeLight:
		switch fieldID {
		case FieldState:
			return newMetricMap(metrics.MetricTypeBinary, "", false)
		case FieldBrightness:
			return newMetricMap(metrics.MetricTypeNone, "%", false)
		default:
			return newMetricMap(metrics.MetricTypeNone, "", false)
		}

	case EntityTypeTextSensor:
		return newMetricMap(metrics.MetricTypeNone, "", true)

	case EntityTypeSensor:
		return newMetricMap(metrics.MetricTypeGaugeFloat, "", true)

	}

	return newMetricMap(metrics.MetricTypeNone, "", true)
}
