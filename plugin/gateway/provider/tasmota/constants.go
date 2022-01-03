package tasmota

import (
	"fmt"

	mtsTY "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
)

// tasmota message data
// topic/node-id/command
// node-id example: tasmota_49C88D
// prefix can be one of: cmnd, stat, tele
// command examples: STARUS, POWER1, POWER2, etc.,
type message struct {
	Topic   string
	NodeID  string
	Command string
	Payload string
}

func (m *message) toString() string {
	return fmt.Sprintf("%s/%s/%s", m.Topic, m.NodeID, m.Command)
}

const (
	topicTele = "tele"
	topicStat = "stat"
	topicCmnd = "cmnd"

	emptyPayload = ""
)

// static sources
const (
	sourceIDNone    = ""
	sourceIDControl = "Control"
	sourceIDWiFi    = "WiFi"
	sourceIDMemory  = "Memory"
	sourceIDTime    = "Time"
	sourceIDLogging = "Logging"
	sourceIDCounter = "Counter"
	sourceIDAnalog  = "Analog"
)

// tele state ignore fields
var teleStateFieldsIgnore = []string{
	"time",
	"uptime",
	"uptimesec",
	"sleepmode",
	"sleep",
	"mqttcount",
	"channel",
	"ledtable",
}

// WiFi source fields ignore
var wiFiFieldsIgnore = []string{"ap", "linkcount"}

// supported status
var statusSupported = []string{
	cmdStatus,
	cmdStatus1,
	cmdStatus2,
	cmdStatus3,
	cmdStatus4,
	cmdStatus5,
	cmdStatus6,
	cmdStatus7,
	cmdStatus8,
	cmdStatus9,
	cmdStatus10,
	cmdStatus11,
}

// logging ignore fields
var loggingFieldsIgnore = []string{"ssid", "resolution", "setoption"}

// command status types
const (
	cmdStatus   = "STATUS"
	cmdStatus1  = "STATUS1"
	cmdStatus2  = "STATUS2"
	cmdStatus3  = "STATUS3"
	cmdStatus4  = "STATUS4"
	cmdStatus5  = "STATUS5"
	cmdStatus6  = "STATUS6"
	cmdStatus7  = "STATUS7"
	cmdStatus8  = "STATUS8"
	cmdStatus9  = "STATUS9"
	cmdStatus10 = "STATUS10"
	cmdStatus11 = "STATUS11"

	cmdState  = "STATE"
	cmdSensor = "SENSOR"
	cmdResult = "RESULT"

	cmdRestart = "Restart"
	cmdReset   = "Reset"

	cmdInfo1 = "INFO1"
	cmdInfo2 = "INFO2"
	cmdInfo3 = "INFO3"

	cmdLWT = "LWT"
)

const (
	headerDeviceParameters = "StatusPRM"
	headerFirmware         = "StatusFWR"
	headerLogging          = "StatusLOG"
	headerMemory           = "StatusMEM"
	headerNetwork          = "StatusNET"
	headerTime             = "StatusTIM"
	headerSensor           = "StatusSNS"
	headerStatus           = "Status"
	// headerMQTT             = "StatusMQT"
	// headerPowerThresholds = "StatusPTH"
)

const (
	keyFriendlyName    = "FriendlyName"
	keyCounter         = "COUNTER"
	keyAnalog          = "ANALOG"
	keyTemperature     = "Temperature"
	keyHumidity        = "Humidity"
	keyDeWPoint        = "DewPoint"
	keyTemperatureUnit = "TempUnit"
	keyHeap            = "Heap"
	keyRSSI            = "RSSI"
	keySignal          = "Signal"
	keyIPAddress       = "IPAddress"
	keyVersion         = "Version"
	keyCore            = "Core"
	keySDK             = "SDK"
	keyBuildDateTime   = "BuildDateTime"
	keyCPUFrequency    = "CpuFrequency"
	keyHardware        = "Hardware"
	keyOtaURL          = "OtaUrl"
	keyHostname        = "Hostname"
	keyMAC             = "Mac"
	keyPower           = "POWER"
	keyFade            = "Fade"
	keyWifi            = "Wifi"
	keyVoltage         = "Voltage"
	keyCurrent         = "Current"
	keyHSBColor        = "HSBColor"
	keyHSBColor1       = "HSBColor1"
	keyHSBColor2       = "HSBColor2"
	keyHSBColor3       = "HSBColor3"
	keyRestartReason   = "RestartReason"
	keyModule          = "Module"
	keyFallbackTopic   = "FallbackTopic"
	keyGroupTopic      = "GroupTopic"
	keyBoot            = "boot"

	// keyON               = "ON"
	// keyOFF              = "OFF"
	// keyDimmer           = "Dimmer"
	// keyProgramSize      = "ProgramSize"
	// keyFlashChipID      = "FlashChipId"
	// keyFlashFrequency   = "FlashFrequency"
	// keyFlashSize        = "FlashSize"
	// keyProgramFlashSize = "ProgramFlashSize"
	// keyWebServerMode    = "WebServerMode"
)

// this struct used to construct payload metric type and unit
type payloadMetricTypeUnit struct{ Type, Unit string }

// map default metric types unit types for the fields
var metricTypeAndUnit = map[string]payloadMetricTypeUnit{
	keyTemperature: {mtsTY.MetricTypeGaugeFloat, mtsTY.UnitCelsius},
	keyHumidity:    {mtsTY.MetricTypeGaugeFloat, mtsTY.UnitPercent},
	keyDeWPoint:    {mtsTY.MetricTypeGaugeFloat, mtsTY.UnitCelsius},
	keyHeap:        {mtsTY.MetricTypeGauge, mtsTY.UnitNone},
	keyPower:       {mtsTY.MetricTypeBinary, mtsTY.UnitNone},
	keyFade:        {mtsTY.MetricTypeBinary, mtsTY.UnitNone},
	keyRSSI:        {mtsTY.MetricTypeGauge, mtsTY.UnitNone},
	keySignal:      {mtsTY.MetricTypeGauge, mtsTY.UnitNone},
	keyVoltage:     {mtsTY.MetricTypeGaugeFloat, mtsTY.UnitNone},
	keyCurrent:     {mtsTY.MetricTypeGaugeFloat, mtsTY.UnitNone},
}
