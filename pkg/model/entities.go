package model

import "fmt"

// Entities
const (
	EntityGateway  = "gateway"  // keeps gateway config details
	EntityNode     = "node"     // keeps node details
	EntitySensor   = "sensor"   // keeps sensor details
	EntityField    = "field"    // keeps sensor field details and fields from node, like battery, rssi, etc.,
	EntityFirmware = "firmware" // keeps firmware details
)

// Files, directory locations
const (
	DirectoryRoot     = "/tmp/myc"  // root directory for all the files
	DirectoryFirmware = "/firmware" // location to keep firmware files
)

// DirectoryFullPath adds root dir and return full path
func DirectoryFullPath(subDir string) string {
	return fmt.Sprintf("%s%s", DirectoryRoot, subDir)
}
