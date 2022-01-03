package types

// reserved labels
const (
	// resource ids
	LabelGatewayID = "gateway_id"
	LabelNodeID    = "node_id"
	LabelSourceID  = "source_id"
	LabelFieldID   = "field_id"

	// Common labels
	LabelName      = "name"
	LabelTimezone  = "timezone"
	LabelReadOnly  = "readonly"
	LabelWriteOnly = "writeonly"

	// Node specific labels
	LabelNodeSleepNode        = "sleep_node"
	LabelNodeVersion          = "version"
	LabelNodeLibraryVersion   = "library_version"
	LabelNodeAssignedFirmware = "assigned_firmware"
	LabelNodeOTABlockOrder    = "ota_block_order"

	// Field specific labels
	LabelMetricType = "metric_type"
	LabelUnit       = "unit"

	// device specific
	LabelDeviceType = "device_type"
	LabelTraitType  = "trait_type"
	LabelParamType  = "param_type"
)
