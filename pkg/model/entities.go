package model

// Entities
const (
	EntityGateway        = "gateway"         // keeps gateway config details
	EntityNode           = "node"            // keeps node details
	EntitySensor         = "sensor"          // keeps sensor details
	EntitySensorField    = "sensor_field"    // keeps sensor field details and fields from node, like battery, rssi, etc.,
	EntityFirmware       = "firmware"        // keeps firmware details
	EntityUser           = "user"            // keeps user details
	EntityDashboard      = "dashboard"       // keeps dashboard details
	EntityForwardPayload = "forward_payload" // keeps forward payload mapping details
	EntityNotifyHandler  = "notify_handler"  // keeps configurations for notify handlers
	EntityTask           = "task"            // keeps configurations for tasks
	EntityScheduler      = "scheduler"       // keeps configurations for scheduler
)

// Kind types
const (
	KindExportConfig   = "ExportConfig"
	KindExporterConfig = "ExporterConfig"
)

// Entity field keys
const (
	KeyID          = "ID"
	KeyGatewayID   = "GatewayID"
	KeyNodeID      = "NodeID"
	KeySensorID    = "SensorID"
	KeyFieldID     = "FieldID"
	KeyUsername    = "Username"
	KeyEmail       = "Email"
	KeyHandlerType = "Type"
	KeyHandlerName = "Name"
	KeySourceID    = "SourceID"
	KeyEnabled     = "Enabled"
)

// Field names used in entities
const (
	NameType = "type"
)

// keys used in other locations
const (
	KeySelector = "selector"
	KeyTemplate = "template"
	KeyTask     = "task"
	KeyPayload  = "payload"
)
