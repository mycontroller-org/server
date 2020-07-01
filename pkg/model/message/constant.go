package message

// DataTypes of a payload
var DataTypes = []string{
	DataTypeFloat,
	DataTypeInteger,
	DataTypeString,
	DataTypeBoolean,
}

// Data type will be used for metrics
const (
	DataTypeFloat   = "float"
	DataTypeInteger = "integer"
	DataTypeString  = "string"
	DataTypeBoolean = "boolean"
	DataTypeGeo     = "geo"
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
	KeyCmdNodeName                  = "name"
	KeyCmdNodeReboot                = "reboot"
	KeyCmdNodeVersion               = "version"
	KeyCmdNodeLibraryVersion        = "libraryVersion"
	KeyCmdNodeParentID              = "parentId"
	KeyCmdNodeDiscover              = "descover"
	KeyCmdNodeHeartbeat             = "hearbeat"
	KeyCmdNodePing                  = "ping"
	KeyCmdNodeSignalStrength        = "signalStrength"
	KeyCmdNodeBatteryLevel          = "batteryLevel"
	KeyCmdNodePreSleepNotification  = "preSleep"
	KeyCmdNodePostSleepNotification = "postSleep"
)

// Others map known keys
const (
	KeyTopic = "topic"
	KeyQoS   = "qos"
	KeyName  = "name"
)
