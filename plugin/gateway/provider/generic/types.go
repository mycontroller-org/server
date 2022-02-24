package generic

import (
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

const (
	KeyReceivedPayload  = "payload_json"
	KeyReceivedMessages = "messages"
)

// Config of generic provider
type Config struct {
	Type       string          `json:"type"`
	RetryCount int             `json:"retryCount"`
	Script     ScriptFormatter `json:"script"`
	Nodes      cmap.CustomMap  `json:"nodes"`
	Protocol   cmap.CustomMap  `json:"protocol"` // mqtt type will be handled by default mqtt protocol
}

// script used to format
type ScriptFormatter struct {
	OnReceive string `json:"onReceive"`
	OnSend    string `json:"onSend"`
}

// http protocol config
type HttpProtocol struct {
	Type            string                 `json:"type"`
	Headers         map[string]string      `json:"headers"`
	QueryParameters map[string]interface{} `json:"queryParameters"`
	Endpoints       map[string]HttpConfig  `json:"endpoints"`
}

// http config
type HttpConfig struct {
	HttpNode
	Disabled          bool   `json:"disabled"`
	ExecutionInterval string `json:"executionInterval"`
}

// nodes details

// http node config
type HttpNode struct {
	URL             string                 `json:"url"`
	Method          string                 `json:"method"`
	Insecure        bool                   `json:"insecure"`
	Headers         map[string]string      `json:"headers"`
	QueryParameters map[string]interface{} `json:"queryParameters"`
	Body            cmap.CustomMap         `json:"body"`
	ResponseCode    int                    `json:"responseCode"`
	Script          string                 `json:"script"`
	IncludeGlobal   bool                   `json:"includeGlobal"`
}

// mqtt node config
type MqttNode struct {
	Topic  string `json:"topic"`
	QoS    string `json:"qos"`
	Script string `json:"script"`
}
