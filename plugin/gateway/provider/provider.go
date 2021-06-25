package provider

import (
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
)

// Provider interface
type Provider interface {
	Start(messageReceiveFunc func(rawMsg *msgML.RawMessage) error) error   // start the provider
	Close() error                                                          // close the provider connection
	Post(message *msgML.Message) error                                     // post a message to the provider
	ProcessReceived(rawMesage *msgML.RawMessage) ([]*msgML.Message, error) // process the received messages
}

// Providers list
const (
	TypeMySensorsV2      = "mysensors_v2"
	TypePhilipsHue       = "philips_hue"
	TypeSystemMonitoring = "system_monitoring"
	TypeTasmota          = "tasmota"
	TypeEsphome          = "esphome"
	TypeCustom           = "custom"
)
