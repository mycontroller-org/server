package types

// reserved labels
// there is common prefix in all resources
const (
	// resource ids
	LabelGatewayID = "gateway_id"
	LabelNodeID    = "node_id"
	LabelSourceID  = "source_id"
	LabelFieldID   = "field_id"

	// Common labels
	LabelName      = "name"
	LabelTimezone  = "timezone"
	LabelReadOnly  = "read_only"
	LabelWriteOnly = "write_only"

	// Node specific labels
	LabelNodeSleepNode          = "sleep_node"           // messages always will be kept in sleep queue
	LabelNodeSleepQueueDisabled = "sleep_queue_disabled" // failed messages and messages will not be kept in sleep queue
	LabelNodeVersion            = "version"              // version of the application
	LabelNodeLibraryVersion     = "library_version"      // version of the underlying library
	LabelNodeAssignedFirmware   = "assigned_firmware"    // assigned firmware of the node
	LabelNodeOTABlockOrder      = "ota_block_order"      // used on OTA
	LabelNodeInactiveDuration   = "inactive_duration"    // mark the node as DOWN, if there is no message on the specified period

	// Field specific labels
	LabelMetricType = "metric_type"
	LabelUnit       = "unit"

	// device specific
	LabelDeviceType = "device_type"
	LabelTraitType  = "trait_type"
	LabelParamType  = "param_type"

	// used in virtual device
	LabelRoom = "room"
)
