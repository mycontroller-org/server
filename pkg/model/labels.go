package model

// reserved labels
const (
	// resource ids
	LabelGatewayID = "gateway_id"
	LabelNodeID    = "node_id"
	LabelSensorID  = "sensor_id"
	LabelFieldID   = "field_id"

	// Common labels
	LabelName     = "name"
	LabelTimezone = "timezone"
	LabelReadOnly = "readonly"

	// Node specific labels
	LabelNodeIsSleepingNode   = "is_sleeping_node"
	LabelNodeVersion          = "version"
	LabelNodeLibraryVersion   = "library_version"
	LabelNodeAssignedFirmware = "assigned_firmware"

	// Field specific labels
	LabelMetricType = "metric_type"
	LabelUnit       = "unit"

	// Widget label
	LabelWideget = "wideget"
)

// UI widegets
const (
	WidegetSwitch = "switch"
	WidgetGauge   = "gauge"
	WidegetRGB    = "rgb"
	WidegetRGBW   = "rgbw"
	WidegetRGBWC  = "rgbwc"
)
