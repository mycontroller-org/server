package model

// Entities
const (
	EntityGateway              = "gateway"               // keeps gateway config details
	EntityNode                 = "node"                  // keeps node details
	EntitySensor               = "sensor"                // keeps sensor details
	EntitySensorField          = "sensor_field"          // keeps sensor field details and fields from node, like battery, rssi, etc.,
	EntityFirmware             = "firmware"              // keeps firmware details
	EntityUser                 = "user"                  // keeps user details
	EntityDashboard            = "dashboard"             // keeps dashboard details
	EntityForwardPayload       = "forward_payload"       // keeps forward payload mapping details
	EntityNotificationHandlers = "notification_handlers" // keeps configurations for notification handlers
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
	KeyServiceType = "Type"
	KeyServiceName = "Name"
	KeySourceID    = "SourceID"
	KeyEnabled     = "Enabled"
)

// Field names used in entities
const (
	NameType = "type"
)
