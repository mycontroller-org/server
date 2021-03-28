package tasmota

import (
	"fmt"

	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
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
	headerMQTT             = "StatusMQT"
	headerTime             = "StatusTIM"
	headerSensor           = "StatusSNS"
	headerPowerThresholds  = "StatusPTH"
	headerStatus           = "Status"
)

const (
	keyFriendlyName     = "FriendlyName"
	keyCounter          = "COUNTER"
	keyAnalog           = "ANALOG"
	keyTemperature      = "Temperature"
	keyHumidity         = "Humidity"
	keyDeWPoint         = "DewPoint"
	keyTemperatureUnit  = "TempUnit"
	keyON               = "ON"
	keyOFF              = "OFF"
	keyHeap             = "Heap"
	keyRSSI             = "RSSI"
	keySignal           = "Signal"
	keyDimmer           = "Dimmer"
	keyIPAddress        = "IPAddress"
	keyVersion          = "Version"
	keyCore             = "Core"
	keySDK              = "SDK"
	keyBuildDateTime    = "BuildDateTime"
	keyCPUFrequency     = "CpuFrequency"
	keyHardware         = "Hardware"
	keyOtaURL           = "OtaUrl"
	keyHostname         = "Hostname"
	keyMAC              = "Mac"
	keyProgramSize      = "ProgramSize"
	keyFlashChipID      = "FlashChipId"
	keyFlashFrequency   = "FlashFrequency"
	keyFlashSize        = "FlashSize"
	keyProgramFlashSize = "ProgramFlashSize"
	keyPower            = "POWER"
	keyFade             = "Fade"
	keyWifi             = "Wifi"
	keyVoltage          = "Voltage"
	keyCurrent          = "Current"
	keyHSBColor         = "HSBColor"
	keyHSBColor1        = "HSBColor1"
	keyHSBColor2        = "HSBColor2"
	keyHSBColor3        = "HSBColor3"
	keyRestartReason    = "RestartReason"
	keyModule           = "Module"
	keyFallbackTopic    = "FallbackTopic"
	keyGroupTopic       = "GroupTopic"
	keyWebServerMode    = "WebServerMode"
	keyBoot             = "boot"
)

// this struct used to construct payload metric type and unit
type payloadMetricTypeUnit struct{ Type, Unit string }

// map default metric types unit types for the fields
var metricTypeAndUnit = map[string]payloadMetricTypeUnit{
	keyTemperature: {mtsml.MetricTypeGaugeFloat, mtsml.UnitCelsius},
	keyHumidity:    {mtsml.MetricTypeGaugeFloat, mtsml.UnitHumidity},
	keyDeWPoint:    {mtsml.MetricTypeGaugeFloat, mtsml.UnitCelsius},
	keyHeap:        {mtsml.MetricTypeGauge, mtsml.UnitNone},
	keyPower:       {mtsml.MetricTypeBinary, mtsml.UnitNone},
	keyFade:        {mtsml.MetricTypeBinary, mtsml.UnitNone},
	keyRSSI:        {mtsml.MetricTypeGauge, mtsml.UnitNone},
	keySignal:      {mtsml.MetricTypeGauge, mtsml.UnitNone},
	keyVoltage:     {mtsml.MetricTypeGaugeFloat, mtsml.UnitNone},
	keyCurrent:     {mtsml.MetricTypeGaugeFloat, mtsml.UnitNone},
}
