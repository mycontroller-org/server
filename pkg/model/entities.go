package model

// Entities
const (
	EntityGateway     = "gateway"      // keeps gateway config details
	EntityNode        = "node"         // keeps node details
	EntitySensor      = "sensor"       // keeps sensor details
	EntitySensorField = "sensor_field" // keeps sensor field details and fields from node, like battery, rssi, etc.,
	EntityFirmware    = "firmware"     // keeps firmware details
	EntityUser        = "user"         // keeps user details
	EntityKind        = "kind"         // keeps configurations, job details, rules, operations, etc..,
)

// Kind types
const (
	KindExportConfig   = "ExportConfig"
	KindExporterConfig = "ExporterConfig"
)

// Entity field keys
const (
	KeyID        = "ID"
	KeyGatewayID = "GatewayID"
	KeyNodeID    = "NodeID"
	KeySensorID  = "SensorID"
	KeyFieldID   = "FieldID"
	KeyUsername  = "Username"
	KeyEmail     = "Email"
	KeyKindType  = "Type"
	KeyKindName  = "Name"
)

// Field names used in entities
const (
	NameType = "type"
)
