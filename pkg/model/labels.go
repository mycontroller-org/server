package model

// reserved labels
const (
	// resource ids
	LabelGatewayID = "gateway_id"
	LabelNodeID    = "node_id"
	LabelSourceID  = "source_id"
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
	LabelNodeBootloader       = "bootloader"

	// Field specific labels
	LabelMetricType = "metric_type"
	LabelUnit       = "unit"
)
