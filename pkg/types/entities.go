package types

// Entities
const (
	EntityGateway          = "gateway"           // keeps gateway config details
	EntityNode             = "node"              // keeps node details
	EntitySource           = "source"            // keeps source details
	EntityField            = "field"             // keeps field details from source and fields from node, like battery, rssi, etc.,
	EntityFirmware         = "firmware"          // keeps firmware details
	EntityUser             = "user"              // keeps user details
	EntityDashboard        = "dashboard"         // keeps dashboard details
	EntityForwardPayload   = "forward_payload"   // keeps forward payload mapping details
	EntityHandler          = "handler"           // keeps configurations for handlers
	EntityTask             = "task"              // keeps configurations for tasks
	EntitySchedule         = "schedule"          // keeps configurations for schedules
	EntitySettings         = "settings"          // settings of the system
	EntityDataRepository   = "data_repository"   // holds user data, can be used across
	EntityVirtualDevice    = "virtual_device"    // holds virtual devices
	EntityVirtualAssistant = "virtual_assistant" // holds virtual assistants
	EntityServiceToken     = "service_token"     // holds service token
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
	KeyUserID       = "UserID"
	KeyTokenID      = "Token.ID"
	KeyEmail        = "Email"
	KeyHandlerType  = "Type"
	KeyHandlerName  = "Name"
	KeyEnabled      = "Enabled"
	KeyDisabled     = "Disabled"
	KeyScheduleType = "Type"
	KeyType         = "Type"
	KeySrcFieldID   = "SrcFieldID"
	KeyName         = "Name"
	KeyLocation     = "Location"
)

// Field names used in entities
const (
	NameType = "type"
)

// keys used in other locations
const (
	KeyKeyPath     = "keyPath"
	KeyTemplate    = "template"
	KeyTask        = "task"
	KeySchedule    = "schedule"
	KeyEventType   = "eventType"
	KeyTaskEvent   = "taskEvent"
	KeyPayload     = "payload"
	KeyValue       = "value"
	KeyMetricTypes = "metricTypes"
	KeyOthers      = "others"
	KeyLabels      = "labels"
	KeyUnits       = "units"
)
