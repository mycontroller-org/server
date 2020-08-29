package mysensors

import (
	"fmt"

	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

// Config key, will be used in gateway provider config
const (
	KeyProviderUnitsConfig = "unitsConfig"
	idBroadcast            = "255"

	// Others data key
	OthersKeyNodeID = "nodeID"
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
	keyType       = "ms_type"
	KeyTypeString = "ms_type_string"
	keyNodeType   = "ms_node_type"
	keyNodeID     = "ms_node_id"
	keySensorID   = "ms_sensor_id"
)

// Constants of MySensors
const (
	SerialMessageSplitter = '\n'
)

// mysensors message data
// node-id/child-sensor-id/command/ack/type
type message struct {
	NodeID   string
	SensorID string
	Command  string
	Ack      string
	Type     string
	Payload  string
}

func (ms *message) toRaw(isMQTT bool) string {
	// raw message format
	// node-id;child-sensor-id;command;ack;type;payload
	if ms.NodeID == "" {
		ms.NodeID = idBroadcast
	}
	if ms.SensorID == "" {
		ms.SensorID = idBroadcast
	}
	if isMQTT {
		return fmt.Sprintf("%s/%s/%s/%s/%s", ms.NodeID, ms.SensorID, ms.Command, ms.Ack, ms.Type)
	}
	return fmt.Sprintf("%s;%s;%s;%s;%s;%s\n", ms.NodeID, ms.SensorID, ms.Command, ms.Ack, ms.Type, ms.Payload)
}

// Command in map
var cmdMapIn = map[string]string{
	"0": msgml.TypePresentation,
	"1": msgml.TypeSet,
	"2": msgml.TypeRequest,
	"3": msgml.TypeInternal,
	"4": msgml.TypeStream,
}

// Command out map
var cmdMapOut = map[string]string{
	msgml.TypePresentation: "0",
	msgml.TypeSet:          "1",
	msgml.TypeRequest:      "2",
	msgml.TypeInternal:     "3",
	msgml.TypeStream:       "4",
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

var cmdSetReqFieldMapIn = map[string]string{
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

// PayloadMetricTypeUnit struct
type PayloadMetricTypeUnit struct{ Type, Unit string }

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

var metricTypeUnit = map[string]PayloadMetricTypeUnit{
	"V_TEMP":               {fml.MetricTypeGaugeFloat, unitCelsius},
	"V_HUM":                {fml.MetricTypeGaugeFloat, unitHumidity},
	"V_STATUS":             {fml.MetricTypeBinary, unitNone},
	"V_PERCENTAGE":         {fml.MetricTypeGaugeFloat, unitPercent},
	"V_PRESSURE":           {fml.MetricTypeGaugeFloat, unitNone},
	"V_FORECAST":           {fml.MetricTypeGaugeFloat, unitNone},
	"V_RAIN":               {fml.MetricTypeGaugeFloat, unitNone},
	"V_RAINRATE":           {fml.MetricTypeGaugeFloat, unitNone},
	"V_WIND":               {fml.MetricTypeGaugeFloat, unitNone},
	"V_GUST":               {fml.MetricTypeGaugeFloat, unitNone},
	"V_DIRECTION":          {fml.MetricTypeGaugeFloat, unitNone},
	"V_UV":                 {fml.MetricTypeGaugeFloat, unitNone},
	"V_WEIGHT":             {fml.MetricTypeGaugeFloat, unitNone},
	"V_DISTANCE":           {fml.MetricTypeGaugeFloat, unitNone},
	"V_IMPEDANCE":          {fml.MetricTypeGaugeFloat, unitNone},
	"V_ARMED":              {fml.MetricTypeBinary, unitNone},
	"V_TRIPPED":            {fml.MetricTypeBinary, unitNone},
	"V_WATT":               {fml.MetricTypeGaugeFloat, unitNone},
	"V_KWH":                {fml.MetricTypeGaugeFloat, unitNone},
	"V_SCENE_ON":           {fml.MetricTypeNone, unitNone},
	"V_SCENE_OFF":          {fml.MetricTypeNone, unitNone},
	"V_HVAC_FLOW_STATE":    {fml.MetricTypeGaugeFloat, unitNone},
	"V_HVAC_SPEED":         {fml.MetricTypeGaugeFloat, unitNone},
	"V_LIGHT_LEVEL":        {fml.MetricTypeGaugeFloat, unitPercent},
	"V_VAR1":               {fml.MetricTypeNone, unitNone},
	"V_VAR2":               {fml.MetricTypeNone, unitNone},
	"V_VAR3":               {fml.MetricTypeNone, unitNone},
	"V_VAR4":               {fml.MetricTypeNone, unitNone},
	"V_VAR5":               {fml.MetricTypeNone, unitNone},
	"V_UP":                 {fml.MetricTypeBinary, unitNone},
	"V_DOWN":               {fml.MetricTypeBinary, unitNone},
	"V_STOP":               {fml.MetricTypeBinary, unitNone},
	"V_IR_SEND":            {fml.MetricTypeNone, unitNone},
	"V_IR_RECEIVE":         {fml.MetricTypeNone, unitNone},
	"V_FLOW":               {fml.MetricTypeGaugeFloat, unitNone},
	"V_VOLUME":             {fml.MetricTypeGaugeFloat, unitNone},
	"V_LOCK_STATUS":        {fml.MetricTypeBinary, unitNone},
	"V_LEVEL":              {fml.MetricTypeGaugeFloat, unitNone},
	"V_VOLTAGE":            {fml.MetricTypeGaugeFloat, unitVoltage},
	"V_CURRENT":            {fml.MetricTypeGaugeFloat, unitAmpere},
	"V_RGB":                {fml.MetricTypeNone, unitNone},
	"V_RGBW":               {fml.MetricTypeNone, unitNone},
	"V_ID":                 {fml.MetricTypeNone, unitNone},
	"V_UNIT_PREFIX":        {fml.MetricTypeNone, unitNone},
	"V_HVAC_SETPOINT_COOL": {fml.MetricTypeGaugeFloat, unitNone},
	"V_HVAC_SETPOINT_HEAT": {fml.MetricTypeGaugeFloat, unitNone},
	"V_HVAC_FLOW_MODE":     {fml.MetricTypeNone, unitNone},
	"V_TEXT":               {fml.MetricTypeNone, unitNone},
	"V_CUSTOM":             {fml.MetricTypeNone, unitNone},
	"V_POSITION":           {fml.MetricTypeNone, unitNone},
	"V_IR_RECORD":          {fml.MetricTypeNone, unitNone},
	"V_PH":                 {fml.MetricTypeGaugeFloat, unitNone},
	"V_ORP":                {fml.MetricTypeGaugeFloat, unitNone},
	"V_EC":                 {fml.MetricTypeGaugeFloat, unitNone},
	"V_VAR":                {fml.MetricTypeGaugeFloat, unitNone},
	"V_VA":                 {fml.MetricTypeGaugeFloat, unitNone},
	"V_POWER_FACTOR":       {fml.MetricTypeGaugeFloat, unitNone},
}

var cmdInternalTypeMapIn = map[string]string{
	"0":  "I_BATTERY_LEVEL",
	"1":  "I_TIME",
	"2":  "I_VERSION",
	"3":  "I_ID_REQUEST",
	"4":  "I_ID_RESPONSE",
	"5":  "I_INCLUSION_MODE",
	"6":  "I_CONFIG",
	"7":  "I_FIND_PARENT",
	"8":  "I_FIND_PARENT_RESPONSE",
	"9":  "I_LOG_MESSAGE",
	"10": "I_CHILDREN",
	"11": "I_SKETCH_NAME",
	"12": "I_SKETCH_VERSION",
	"13": "I_REBOOT",
	"14": "I_GATEWAY_READY",
	"15": "I_SIGNING_PRESENTATION",
	"16": "I_NONCE_REQUEST",
	"17": "I_NONCE_RESPONSE",
	"18": "I_HEARTBEAT_REQUEST",
	"19": "I_PRESENTATION",
	"20": "I_DISCOVER_REQUEST",
	"21": "I_DISCOVER_RESPONSE",
	"22": "I_HEARTBEAT_RESPONSE",
	"23": "I_LOCKED",
	"24": "I_PING",
	"25": "I_PONG",
	"26": "I_REGISTRATION_REQUEST",
	"27": "I_REGISTRATION_RESPONSE",
	"28": "I_DEBUG",
	"29": "I_SIGNAL_REPORT_REQUEST",
	"30": "I_SIGNAL_REPORT_REVERSE",
	"31": "I_SIGNAL_REPORT_RESPONSE",
	"32": "I_PRE_SLEEP_NOTIFICATION",
	"33": "I_POST_SLEEP_NOTIFICATION",
}

var cmdInternalFieldMap = map[string]string{
	"I_BATTERY_LEVEL":          fml.FieldBatteryLevel,
	"I_VERSION":                fml.FieldLibraryVersion,
	"I_FIND_PARENT_RESPONSE":   fml.FieldParentID,
	"I_SKETCH_NAME":            fml.FieldName,
	"I_SKETCH_VERSION":         fml.FieldVersion,
	"I_LOCKED":                 fml.FieldLocked,
	"I_SIGNAL_REPORT_RESPONSE": fml.FieldSignalStrength,
}

type nodeInternalCmd struct {
	Field      string
	callBackFn func() *msgml.RawMessage
}

const (
	nodeIDRequest  = "node_id_request"
	nodeIDResponse = "node_id_response"
)

var localHandlerMapIn = map[string]func(gw *gwml.Config, ms message) *message{
	"1": timeHandler,      // I_TIME
	"2": idRequestHandler, // I_ID_REQUEST
	"6": configHandler,    // I_CONFIG
}
