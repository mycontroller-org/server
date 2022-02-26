package generic

import (
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
)

// Config of generic provider
type Config struct {
	Type       string          `json:"type"`
	RetryCount int             `json:"retryCount"`
	Script     ScriptFormatter `json:"script"`
	Protocol   cmap.CustomMap  `json:"protocol"` // mqtt type will be handled by default mqtt protocol
}

// script used to format
type ScriptFormatter struct {
	OnReceive string `json:"onReceive"`
	OnSend    string `json:"onSend"`
}

// Generic protocol
type GenericProtocol interface {
	Post(rawMsg *msgTY.Message) error // post a message on a specified protocol
	Close() error                     // close the protocol connection
}

const (
	ScriptKeyDataIn  = "dataIn"
	ScriptKeyDataOut = "dataOut"
)
