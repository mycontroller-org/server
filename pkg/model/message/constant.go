package message

// PayloadTypes of a payload
var PayloadTypes = []string{
	PayloadTypeFloat,
	PayloadTypeInteger,
	PayloadTypeString,
	PayloadTypeBoolean,
}

// Payload type will be used for metrics
const (
	PayloadTypeFloat   = "float"
	PayloadTypeInteger = "integer"
	PayloadTypeString  = "string"
	PayloadTypeBoolean = "boolean"
	PayloadTypeGeo     = "geo"
)

// Units of a payload
var Units = []string{
	UnitNone,
	UnitCentigrade,
	UnitFahrenheit,
	UnitVolt,
	UnitMillivolt,
	UnitMicrovolt,
	UnitAmp,
	UnitMilliamp,
	UnitMicroamp,
	UnitHertz,
	UnitPercentage,
	UnitOhm,
}

// Unit types
const (
	UnitNone       = ""
	UnitCentigrade = "°C"
	UnitFahrenheit = "°F"
	UnitVolt       = "V"
	UnitMillivolt  = "mV"
	UnitMicrovolt  = "µV"
	UnitAmp        = "A"
	UnitMilliamp   = "mA"
	UnitMicroamp   = "µA"
	UnitHertz      = "Hz"
	UnitPercentage = "%"
	UnitOhm        = "Ω"
)

// Commands slice
var Commands = []string{
	CommandSet,
	CommandRequest,
}

// Commands
const (
	CommandNone    = ""
	CommandSet     = "set"
	CommandRequest = "request"
	CommandNode    = "node"
	CommandSensor  = "sensor"
	CommandStream  = "stream"
)

// Node command options
const (
	KeySubCmdName                  = "name"
	KeySubCmdReboot                = "reboot"
	KeySubCmdVersion               = "version"
	KeySubCmdLibraryVersion        = "libraryVersion"
	KeySubCmdParentID              = "parentId"
	KeySubCmdDiscover              = "descover"
	KeySubCmdHeartbeat             = "hearbeat"
	KeySubCmdPing                  = "ping"
	KeySubCmdSignalStrength        = "signalStrength"
	KeySubCmdBatteryLevel          = "batteryLevel"
	KeySubCmdPreSleepNotification  = "preSleep"
	KeySubCmdPostSleepNotification = "postSleep"
)

// Others map known keys
const (
	KeyTopic = "topic"
	KeyQoS   = "qos"
	KeyName  = "name"
)
