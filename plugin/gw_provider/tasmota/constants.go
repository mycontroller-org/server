package tasmota

import (
	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metric"
)

// tasmota message data
// topic/node-id/command
// node-id example: tasmota_49C88D
// prefix can be one of: cmnd, stat, tele
// command examples: STARUS, POWER1, POWER2, etc.,
type message struct {
	NodeID  string
	Topic   string
	Command string
}

// example: jktasmota/stat/tasmota_49C88D/STATUS11
func (m *message) toMessage(topic string) {

}

var cmdMapForRx = map[string]string{}

const (
	topicTele = "tele"
	topicStat = "stat"
	topicCmnd = "cmnd"
)

const (
	sensorIDNone = ""
)

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
)

var cmdWithHeader = []string{
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

// static sensors

const (
	sensorLogging = "Logging"
	sensorMemory  = "Memory"
	sensorCounter = "Counter"
	sensorAnalog  = "Analog"
	sensorPower   = "Power"
)

const (
	keyFriendlyName    = "FriendlyName"
	keyCounter         = "COUNTER"
	keyAnalog          = "ANALOG"
	keyTemperature     = "Temperature"
	keyHumidity        = "Humidity"
	keyDeWPoint        = "DewPoint"
	keyTemperatureUnit = "TempUnit"
	keyON              = "ON"
	keyOFF             = "OFF"
	keyHeap            = "Heap"
	keyRSSI            = "RSSI"
	keySignal          = "Signal"
	keyDimmer          = "Dimmer"
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
)

// this struct used to construct payload metric type and unit
type payloadMetricTypeUnit struct{ Type, Unit string }

// map default metric types unit types for the fields
var metricTypeAndUnit = map[string]payloadMetricTypeUnit{
	keyTemperature: {mtrml.MetricTypeGaugeFloat, mtrml.UnitCelsius},
	keyHumidity:    {mtrml.MetricTypeGaugeFloat, mtrml.UnitHumidity},
	keyDeWPoint:    {mtrml.MetricTypeGaugeFloat, mtrml.UnitCelsius},
	keyHeap:        {mtrml.MetricTypeGauge, mtrml.UnitNone},
}

// Labels used on this provider
const (
	LabelFirmwareTypeID    = "ms_firmware_type_id"    // MySensors firmware type id
	LabelFirmwareVersionID = "ms_firmware_version_id" // MySensors firmware version id
)
