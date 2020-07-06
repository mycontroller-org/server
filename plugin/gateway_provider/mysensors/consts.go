package mysensors

import (
	"fmt"

	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
)

// Config key, will be used in gateway provider config
const (
	KeyProviderUnitsConfig = "unitsConfig"
)

// Command types and value
const (
	CmdPresentation = "0"
	CmdSet          = "1"
	CmdReq          = "2"
	CmdInternal     = "3"
	CmdStream       = "4"
)

// internal references
const (
	keyCmdType       = "type"
	keyCmdTypeString = "typeString"
	keyNodeType      = "nodeType"
)

// Constants of MySensors
const (
	SerialMessageSplitter = '\n'
)

// mysensors message data
// node-id/child-sensor-id/command/ack/type
type myMessage struct {
	NodeID   string
	SensorID string
	Command  string
	Ack      string
	Type     string
	Payload  string
}

func (ms *myMessage) getID() string {
	return fmt.Sprintf("%s%s%s%s", ms.NodeID, ms.SensorID, ms.Command, ms.Type)
}

func (ms *myMessage) toRaw(isMQTT bool) string {
	// raw message format
	// node-id;child-sensor-id;command;ack;type;payload
	if ms.NodeID == "" {
		ms.NodeID = "255"
	}
	if ms.SensorID == "" {
		ms.SensorID = "255"
	}
	if isMQTT {
		return fmt.Sprintf("%s/%s/%s/%s/%s", ms.NodeID, ms.SensorID, ms.Command, ms.Ack, ms.Type)
	}
	return fmt.Sprintf("%s;%s;%s;%s;%s;%s\n", ms.NodeID, ms.SensorID, ms.Command, ms.Ack, ms.Type, ms.Payload)
}

// Command in map
var cmdMapIn = map[string]string{
	"0": msg.CommandSensor,
	"1": msg.CommandSet,
	"2": msg.CommandRequest,
	"3": msg.CommandNode,
	"4": msg.CommandStream,
}

// Command out map
var cmdMapOut = map[string]string{
	msg.CommandSensor:  "0",
	msg.CommandSet:     "1",
	msg.CommandRequest: "2",
	msg.CommandNode:    "3",
	msg.CommandStream:  "4",
}

var cmdPresentationTypeMapIn = map[string]string{
	"0":  "S_DOOR",
	"1":  "S_MOTION",
	"2":  "S_SMOKE",
	"3":  "S_BINARY",
	"4":  "S_DIMMER",
	"5":  "S_COVER",
	"6":  "S_TEMP",
	"7":  "S_HUM",
	"8":  "S_BARO",
	"9":  "S_WIND",
	"10": "S_RAIN",
	"11": "S_UV",
	"12": "S_WEIGHT",
	"13": "S_POWER",
	"14": "S_HEATER",
	"15": "S_DISTANCE",
	"16": "S_LIGHT_LEVEL",
	"17": "S_ARDUINO_NODE",
	"18": "S_ARDUINO_REPEATER_NODE",
	"19": "S_LOCK",
	"20": "S_IR",
	"21": "S_WATER",
	"22": "S_AIR_QUALITY",
	"23": "S_CUSTOM",
	"24": "S_DUST",
	"25": "S_SCENE_CONTROLLER",
	"26": "S_RGB_LIGHT",
	"27": "S_RGBW_LIGHT",
	"28": "S_COLOR_SENSOR",
	"29": "S_HVAC",
	"30": "S_MULTIMETER",
	"31": "S_SPRINKLER",
	"32": "S_WATER_LEAK",
	"33": "S_SOUND",
	"34": "S_VIBRATION",
	"35": "S_MOISTURE",
	"36": "S_INFO",
	"37": "S_GAS",
	"38": "S_GPS",
	"39": "S_WATER_QUALITY",
}

var cmdSetReqTypeMapIn = map[string]string{
	"0":  "V_TEMP",
	"1":  "V_HUM",
	"2":  "V_STATUS",
	"3":  "V_PERCENTAGE",
	"4":  "V_PRESSURE",
	"5":  "V_FORECAST",
	"6":  "V_RAIN",
	"7":  "V_RAINRATE",
	"8":  "V_WIND",
	"9":  "V_GUST",
	"10": "V_DIRECTION",
	"11": "V_UV",
	"12": "V_WEIGHT",
	"13": "V_DISTANCE",
	"14": "V_IMPEDANCE",
	"15": "V_ARMED",
	"16": "V_TRIPPED",
	"17": "V_WATT",
	"18": "V_KWH",
	"19": "V_SCENE_ON",
	"20": "V_SCENE_OFF",
	"21": "V_HVAC_FLOW_STATE",
	"22": "V_HVAC_SPEED",
	"23": "V_LIGHT_LEVEL",
	"24": "V_VAR1",
	"25": "V_VAR2",
	"26": "V_VAR3",
	"27": "V_VAR4",
	"28": "V_VAR5",
	"29": "V_UP",
	"30": "V_DOWN",
	"31": "V_STOP",
	"32": "V_IR_SEND",
	"33": "V_IR_RECEIVE",
	"34": "V_FLOW",
	"35": "V_VOLUME",
	"36": "V_LOCK_STATUS",
	"37": "V_LEVEL",
	"38": "V_VOLTAGE",
	"39": "V_CURRENT",
	"40": "V_RGB",
	"41": "V_RGBW",
	"42": "V_ID",
	"43": "V_UNIT_PREFIX",
	"44": "V_HVAC_SETPOINT_COOL",
	"45": "V_HVAC_SETPOINT_HEAT",
	"46": "V_HVAC_FLOW_MODE",
	"47": "V_TEXT",
	"48": "V_CUSTOM",
	"49": "V_POSITION",
	"50": "V_IR_RECORD",
	"51": "V_PH",
	"52": "V_ORP",
	"53": "V_EC",
	"54": "V_VAR",
	"55": "V_VA",
	"56": "V_POWER_FACTOR",
}

type dataTypeUnit struct{ Type, Unit string }

// take unit details from
// https://github.com/grafana/grafana/blob/v6.7.1/packages/grafana-data/src/valueFormats/categories.ts#L23

// id map
const (
	unitNone     = "none"
	unitCelsius  = "celsius"
	unitHumidity = "humidity"
	unitPercent  = "percent"
	unitVoltage  = "volt"
	unitAmpere   = "amp"
)

var metricUnit = map[string]dataTypeUnit{
	"V_TEMP":               {msg.DataTypeFloat, unitCelsius},
	"V_HUM":                {msg.DataTypeFloat, unitHumidity},
	"V_STATUS":             {msg.DataTypeBoolean, unitNone},
	"V_PERCENTAGE":         {msg.DataTypeFloat, unitPercent},
	"V_PRESSURE":           {msg.DataTypeFloat, unitNone},
	"V_FORECAST":           {msg.DataTypeFloat, unitNone},
	"V_RAIN":               {msg.DataTypeFloat, unitNone},
	"V_RAINRATE":           {msg.DataTypeFloat, unitNone},
	"V_WIND":               {msg.DataTypeFloat, unitNone},
	"V_GUST":               {msg.DataTypeFloat, unitNone},
	"V_DIRECTION":          {msg.DataTypeFloat, unitNone},
	"V_UV":                 {msg.DataTypeFloat, unitNone},
	"V_WEIGHT":             {msg.DataTypeFloat, unitNone},
	"V_DISTANCE":           {msg.DataTypeFloat, unitNone},
	"V_IMPEDANCE":          {msg.DataTypeFloat, unitNone},
	"V_ARMED":              {msg.DataTypeBoolean, unitNone},
	"V_TRIPPED":            {msg.DataTypeBoolean, unitNone},
	"V_WATT":               {msg.DataTypeFloat, unitNone},
	"V_KWH":                {msg.DataTypeFloat, unitNone},
	"V_SCENE_ON":           {msg.DataTypeString, unitNone},
	"V_SCENE_OFF":          {msg.DataTypeString, unitNone},
	"V_HVAC_FLOW_STATE":    {msg.DataTypeFloat, unitNone},
	"V_HVAC_SPEED":         {msg.DataTypeFloat, unitNone},
	"V_LIGHT_LEVEL":        {msg.DataTypeFloat, unitPercent},
	"V_VAR1":               {msg.DataTypeString, unitNone},
	"V_VAR2":               {msg.DataTypeString, unitNone},
	"V_VAR3":               {msg.DataTypeString, unitNone},
	"V_VAR4":               {msg.DataTypeString, unitNone},
	"V_VAR5":               {msg.DataTypeString, unitNone},
	"V_UP":                 {msg.DataTypeBoolean, unitNone},
	"V_DOWN":               {msg.DataTypeBoolean, unitNone},
	"V_STOP":               {msg.DataTypeBoolean, unitNone},
	"V_IR_SEND":            {msg.DataTypeString, unitNone},
	"V_IR_RECEIVE":         {msg.DataTypeString, unitNone},
	"V_FLOW":               {msg.DataTypeFloat, unitNone},
	"V_VOLUME":             {msg.DataTypeFloat, unitNone},
	"V_LOCK_STATUS":        {msg.DataTypeBoolean, unitNone},
	"V_LEVEL":              {msg.DataTypeFloat, unitNone},
	"V_VOLTAGE":            {msg.DataTypeFloat, unitVoltage},
	"V_CURRENT":            {msg.DataTypeFloat, unitAmpere},
	"V_RGB":                {msg.DataTypeString, unitNone},
	"V_RGBW":               {msg.DataTypeString, unitNone},
	"V_ID":                 {msg.DataTypeString, unitNone},
	"V_UNIT_PREFIX":        {msg.DataTypeString, unitNone},
	"V_HVAC_SETPOINT_COOL": {msg.DataTypeFloat, unitNone},
	"V_HVAC_SETPOINT_HEAT": {msg.DataTypeFloat, unitNone},
	"V_HVAC_FLOW_MODE":     {msg.DataTypeString, unitNone},
	"V_TEXT":               {msg.DataTypeString, unitNone},
	"V_CUSTOM":             {msg.DataTypeString, unitNone},
	"V_POSITION":           {msg.DataTypeString, unitNone},
	"V_IR_RECORD":          {msg.DataTypeString, unitNone},
	"V_PH":                 {msg.DataTypeFloat, unitNone},
	"V_ORP":                {msg.DataTypeFloat, unitNone},
	"V_EC":                 {msg.DataTypeFloat, unitNone},
	"V_VAR":                {msg.DataTypeFloat, unitNone},
	"V_VA":                 {msg.DataTypeFloat, unitNone},
	"V_POWER_FACTOR":       {msg.DataTypeFloat, unitNone},
}

var cmdInternalTypeMapOut = map[string]string{
	"0":  msg.KeyCmdNodeBatteryLevel,
	"7":  msg.KeyCmdNodeParentID,
	"13": msg.KeyCmdNodeReboot,
	"18": msg.KeyCmdNodeHeartbeat,
	"20": msg.KeyCmdNodeDiscover,
	"24": msg.KeyCmdNodePing,
}

var cmdInternalTypeMapIn = map[string]string{
	"0":  msg.KeyCmdNodeBatteryLevel,
	"2":  msg.KeyCmdNodeLibraryVersion,
	"8":  msg.KeyCmdNodeParentID,
	"11": msg.KeyCmdNodeName,
	"12": msg.KeyCmdNodeVersion,
	"13": msg.KeyCmdNodeReboot,
	"22": msg.KeyCmdNodeHeartbeat,
	"21": msg.KeyCmdNodeDiscover,
	"25": msg.KeyCmdNodePing,
	"31": msg.KeyCmdNodeSignalStrength,
	"32": msg.KeyCmdNodePreSleepNotification,
	"33": msg.KeyCmdNodePostSleepNotification,
}

var localHandlerMapIn = map[string]func(gw *ml.GatewayConfig, ms myMessage) *myMessage{
	"1": timeHandler,      // I_TIME
	"2": idRequestHandler, // I_ID_REQUEST
	"6": configHandler,    // I_CONFIG
}
