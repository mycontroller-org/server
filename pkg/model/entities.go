package model

// Entities
const (
	EntityGateway        = "gateway"         // keeps gateway config details
	EntityNode           = "node"            // keeps node details
	EntitySource         = "source"          // keeps source details
	EntityField          = "field"           // keeps field details from source and fields from node, like battery, rssi, etc.,
	EntityFirmware       = "firmware"        // keeps firmware details
	EntityUser           = "user"            // keeps user details
	EntityDashboard      = "dashboard"       // keeps dashboard details
	EntityForwardPayload = "forward_payload" // keeps forward payload mapping details
	EntityHandler        = "handler"         // keeps configurations for handlers
	EntityTask           = "task"            // keeps configurations for tasks
	EntityScheduler      = "scheduler"       // keeps configurations for scheduler
	EntitySettings       = "settings"        // settings of the system
	EntityDataRepository = "data_repository" // holds user data, can be used across
)

// Entity field keys
const (
	KeyID           = "ID"
	KeyGatewayID    = "GatewayID"
	KeyNodeID       = "NodeID"
	KeySourceID     = "SourceID"
	KeyFieldID      = "FieldID"
	KeyFieldName    = "FieldName"
	KeyUsername     = "Username"
	KeyEmail        = "Email"
	KeyHandlerType  = "Type"
	KeyHandlerName  = "Name"
	KeyEnabled      = "Enabled"
	KeyScheduleType = "Type"
	KeySrcFieldID   = "SrcFieldID"
)

// Field names used in entities
const (
	NameType = "type"
)

// keys used in other locations
const (
	KeySelector    = "selector"
	KeyTemplate    = "template"
	KeyTask        = "task"
	KeySchedule    = "schedule"
	KeyEventType   = "eventType"
	KeyEvent       = "event"
	KeyPayload     = "payload"
	KeyValue       = "value"
	KeyMetricTypes = "metricTypes"
	KeyOthers      = "others"
	KeyLabels      = "labels"
	KeyUnits       = "units"
)
