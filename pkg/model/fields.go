package model

// Field names
const (
	FieldNone           = ""
	FieldName           = "name"
	FieldParentID       = "parent_id"
	FieldSignalStrength = "signal_strength"
	FieldBatteryLevel   = "battery_level"
	FieldLocked         = "locked"
	FieldHeartbeat      = "heartbeat"
	FieldIPAddress      = "ip_address"
	FieldNodeWebURL     = "node_web_url"
	FieldOTAProgress    = "ota_progress"     // in percentage
	FieldOTARunning     = "ota_running"      // in bool
	FieldOTABlockNumber = "ota_block_number" // current block number
	FieldOTABlockTotal  = "ota_block_total"  // total block count
	FieldOTAStatusOn    = "ota_status_on"    // time
	FieldOTAStartTime   = "ota_start_time"   // start time
	FieldOTAEndTime     = "ota_end_time"     // end time
	FieldOTATimeTaken   = "ota_time_taken"   // time taken to complete the update
)
