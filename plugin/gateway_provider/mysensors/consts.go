package mysensors

import (
	"fmt"

	gwml "github.com/mycontroller-org/backend/pkg/model/gateway"
	msg "github.com/mycontroller-org/backend/pkg/model/message"
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
	KeyCmdTypeString = "typeString"
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
	"0": msg.CommandPresentation,
	"1": msg.CommandSet,
	"2": msg.CommandRequest,
	"3": msg.CommandInternal,
	"4": msg.CommandStream,
}

// Command out map
var cmdMapOut = map[string]string{
	msg.CommandPresentation: "0",
	msg.CommandSet:          "1",
	msg.CommandRequest:      "2",
	msg.CommandInternal:     "3",
	msg.CommandStream:       "4",
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

type PayloadTypeUnit struct{ Type, Unit string }

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

var metricUnit = map[string]PayloadTypeUnit{
	"V_TEMP":               {msg.PayloadTypeFloat, unitCelsius},
	"V_HUM":                {msg.PayloadTypeFloat, unitHumidity},
	"V_STATUS":             {msg.PayloadTypeBoolean, unitNone},
	"V_PERCENTAGE":         {msg.PayloadTypeFloat, unitPercent},
	"V_PRESSURE":           {msg.PayloadTypeFloat, unitNone},
	"V_FORECAST":           {msg.PayloadTypeFloat, unitNone},
	"V_RAIN":               {msg.PayloadTypeFloat, unitNone},
	"V_RAINRATE":           {msg.PayloadTypeFloat, unitNone},
	"V_WIND":               {msg.PayloadTypeFloat, unitNone},
	"V_GUST":               {msg.PayloadTypeFloat, unitNone},
	"V_DIRECTION":          {msg.PayloadTypeFloat, unitNone},
	"V_UV":                 {msg.PayloadTypeFloat, unitNone},
	"V_WEIGHT":             {msg.PayloadTypeFloat, unitNone},
	"V_DISTANCE":           {msg.PayloadTypeFloat, unitNone},
	"V_IMPEDANCE":          {msg.PayloadTypeFloat, unitNone},
	"V_ARMED":              {msg.PayloadTypeBoolean, unitNone},
	"V_TRIPPED":            {msg.PayloadTypeBoolean, unitNone},
	"V_WATT":               {msg.PayloadTypeFloat, unitNone},
	"V_KWH":                {msg.PayloadTypeFloat, unitNone},
	"V_SCENE_ON":           {msg.PayloadTypeString, unitNone},
	"V_SCENE_OFF":          {msg.PayloadTypeString, unitNone},
	"V_HVAC_FLOW_STATE":    {msg.PayloadTypeFloat, unitNone},
	"V_HVAC_SPEED":         {msg.PayloadTypeFloat, unitNone},
	"V_LIGHT_LEVEL":        {msg.PayloadTypeFloat, unitPercent},
	"V_VAR1":               {msg.PayloadTypeString, unitNone},
	"V_VAR2":               {msg.PayloadTypeString, unitNone},
	"V_VAR3":               {msg.PayloadTypeString, unitNone},
	"V_VAR4":               {msg.PayloadTypeString, unitNone},
	"V_VAR5":               {msg.PayloadTypeString, unitNone},
	"V_UP":                 {msg.PayloadTypeBoolean, unitNone},
	"V_DOWN":               {msg.PayloadTypeBoolean, unitNone},
	"V_STOP":               {msg.PayloadTypeBoolean, unitNone},
	"V_IR_SEND":            {msg.PayloadTypeString, unitNone},
	"V_IR_RECEIVE":         {msg.PayloadTypeString, unitNone},
	"V_FLOW":               {msg.PayloadTypeFloat, unitNone},
	"V_VOLUME":             {msg.PayloadTypeFloat, unitNone},
	"V_LOCK_STATUS":        {msg.PayloadTypeBoolean, unitNone},
	"V_LEVEL":              {msg.PayloadTypeFloat, unitNone},
	"V_VOLTAGE":            {msg.PayloadTypeFloat, unitVoltage},
	"V_CURRENT":            {msg.PayloadTypeFloat, unitAmpere},
	"V_RGB":                {msg.PayloadTypeString, unitNone},
	"V_RGBW":               {msg.PayloadTypeString, unitNone},
	"V_ID":                 {msg.PayloadTypeString, unitNone},
	"V_UNIT_PREFIX":        {msg.PayloadTypeString, unitNone},
	"V_HVAC_SETPOINT_COOL": {msg.PayloadTypeFloat, unitNone},
	"V_HVAC_SETPOINT_HEAT": {msg.PayloadTypeFloat, unitNone},
	"V_HVAC_FLOW_MODE":     {msg.PayloadTypeString, unitNone},
	"V_TEXT":               {msg.PayloadTypeString, unitNone},
	"V_CUSTOM":             {msg.PayloadTypeString, unitNone},
	"V_POSITION":           {msg.PayloadTypeString, unitNone},
	"V_IR_RECORD":          {msg.PayloadTypeString, unitNone},
	"V_PH":                 {msg.PayloadTypeFloat, unitNone},
	"V_ORP":                {msg.PayloadTypeFloat, unitNone},
	"V_EC":                 {msg.PayloadTypeFloat, unitNone},
	"V_VAR":                {msg.PayloadTypeFloat, unitNone},
	"V_VA":                 {msg.PayloadTypeFloat, unitNone},
	"V_POWER_FACTOR":       {msg.PayloadTypeFloat, unitNone},
}

var cmdInternalTypeMapOut = map[string]string{
	"0":  msg.SubCmdBatteryLevel,
	"7":  msg.SubCmdParentID,
	"13": msg.SubCmdReboot,
	"18": msg.SubCmdHeartbeat,
	"20": msg.SubCmdDiscover,
	"24": msg.SubCmdPing,
}

var cmdInternalTypeMapIn = map[string]string{
	"0":  msg.SubCmdBatteryLevel,
	"2":  msg.SubCmdLibraryVersion,
	"8":  msg.SubCmdParentID,
	"11": msg.SubCmdName,
	"12": msg.SubCmdVersion,
	"13": msg.SubCmdReboot,
	"22": msg.SubCmdHeartbeat,
	"21": msg.SubCmdDiscover,
	"25": msg.SubCmdPing,
	"31": msg.SubCmdSignalStrength,
	"32": msg.SubCmdPreSleepNotification,
	"33": msg.SubCmdPostSleepNotification,
}

var localHandlerMapIn = map[string]func(gw *gwml.Config, ms myMessage) *myMessage{
	"1": timeHandler,      // I_TIME
	"2": idRequestHandler, // I_ID_REQUEST
	"6": configHandler,    // I_CONFIG
}
