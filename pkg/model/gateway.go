package model

import (
	"time"

	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
)

// Gateway Types
const (
	GatewayTypeMQTT     = "mqtt"
	GatewayTypeSerial   = "serial"
	GatewayTypeEthernet = "ethernet"
)

// AckConfig data
type AckConfig struct {
	Enabled       bool   `json:"enabled"`
	StreamEnabled bool   `json:"streamEnabled"`
	RetryCount    bool   `json:"retryCount"`
	WaitTime      uint64 `json:"waitTime"`
}

// Gateway providers
const (
	GatewayProviderMySensors = "MySensors"
)

// GatewayProvider data
type GatewayProvider struct {
	Type        string                 `json:"type"`
	GatewayType string                 `json:"gatewayType"`
	Config      map[string]interface{} `json:"config"`
}

// GatewayConfigMQTT data
type GatewayConfigMQTT struct {
	Broker    string
	Subscribe string
	Publish   string
	QoS       int
	Username  string
	Password  string `json:"-"`
}

// GatewayConfig entity
type GatewayConfig struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Enabled   bool            `json:"enabled"`
	AckConfig AckConfig       `json:"ackConfig"`
	State     State           `json:"state"`
	Provider  GatewayProvider `json:"providerConfig"`
	LastSeen  time.Time       `json:"lastSeen"`
}

// GatewayProviderTopics to supply to bus
type GatewayProviderTopics struct {
	PostMessage         string
	PostAcknowledgement string
}

// GatewayMessageParser interface for provider
type GatewayMessageParser interface {
	ToRawMessage(message *msg.Wrapper) (*msg.RawMessage, error)
	ToMessage(message *msg.Wrapper) (*msg.Message, error)
}

// Gateway instance
type Gateway interface {
	Close() error
	Write(rawMessage *msg.RawMessage) error
}

// GatewayService details
type GatewayService struct {
	Config  *GatewayConfig
	Parser  GatewayMessageParser
	Topics  GatewayProviderTopics
	Gateway Gateway
}
