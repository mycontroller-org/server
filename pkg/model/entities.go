package model

import "fmt"

// Entities
const (
	EntityGateway     = "gateway"      // keeps gateway config details
	EntityNode        = "node"         // keeps node details
	EntitySensor      = "sensor"       // keeps sensor details
	EntitySensorField = "sensor_field" // keeps sensor field details and fields from node, like battery, rssi, etc.,
	EntityFirmware    = "firmware"     // keeps firmware details
	EntityKind        = "kind"         // keeps configurations, job details, rules, operations, etc..,
)

// Entity field keys
const (
	KeyID        = "ID"
	KeyGatewayID = "GatewayID"
	KeyNodeID    = "NodeID"
	KeySensorID  = "SensorID"
	KeyFieldID   = "FieldID"
	KeyKindType  = "Type"
	KeyKindName  = "Name"
)

// Files, directory locations
const (
	DirectoryRoot           = "/tmp/myc"          // root directory for all the files
	DirectoryFirmware       = "/firmware"         // location to keep firmware files
	DirectoryGatewayRawLogs = "/gateway_raw_logs" // location to keep gateway raw logs
)

// DirectoryFullPath adds root dir and return full path
func DirectoryFullPath(subDir string) string {
	return fmt.Sprintf("%s%s", DirectoryRoot, subDir)
}
