package provider

import (
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

// Provider interface
type Provider interface {
	Start(messageReceiveFunc func(rawMsg *msgml.RawMessage) error) error
	Process(rawMesage *msgml.RawMessage) ([]*msgml.Message, error)
	Post(message *msgml.Message) error
	Close() error
}

// Providers list
const (
	TypeMySensorsV2      = "mysensors_v2"
	TypePhilipsHue       = "philips_hue"
	TypeSystemMonitoring = "system_monitoring"
	TypeTasmota          = "tasmota"
	TypeEsphome          = "esphome"
)
