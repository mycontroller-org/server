package notifyhandler

import (
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// handler types
const (
	TypeNoop       = "noop"
	TypeEmail      = "email"
	TypeTelegram   = "telegram"
	TypeWebhook    = "webhook"
	TypeSMS        = "sms"
	TypePushbullet = "pushbullet"
	TypeResource   = "resource"
)

// handler data types
const (
	DataTypeEmail      = "email"
	DataTypeTelegram   = "telegram"
	DataTypeWebhook    = "webhook"
	DataTypeSMS        = "sms"
	DataTypePushbullet = "pushbullet"
	DataTypeResource   = "resource"
)

// Config model
type Config struct {
	ID          string               `json:"id"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	Labels      cmap.CustomStringMap `json:"labels"`
	Type        string               `json:"type"`
	Spec        cmap.CustomMap       `json:"spec"`
	State       *model.State         `json:"state"`
}

// Clone config
func (hdr *Config) Clone() Config {
	clonedConfig := Config{
		ID:          hdr.ID,
		Description: hdr.Description,
		Enabled:     hdr.Enabled,
		Type:        hdr.Type,
		Labels:      hdr.Labels.Clone(),
		Spec:        hdr.Spec.Clone(),
	}
	return clonedConfig
}

// MessageWrapper to use in bus
// specially used to send data to handlers
type MessageWrapper struct {
	ID   string
	Data map[string]interface{}
}

// GetDataType returns type of the handler
func GetDataType(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.HasPrefix(name, DataTypeEmail):
		return DataTypeEmail
	case strings.HasPrefix(name, DataTypeTelegram):
		return DataTypeTelegram
	case strings.HasPrefix(name, DataTypeWebhook):
		return DataTypeWebhook
	case strings.HasPrefix(name, DataTypeSMS):
		return DataTypeSMS
	case strings.HasPrefix(name, DataTypePushbullet):
		return DataTypePushbullet
	case strings.HasPrefix(name, DataTypeResource):
		return DataTypeResource
	default:
		return ""
	}
}

// GenericData struct
type GenericData struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// ResourceData struct
type ResourceData struct {
	ResourceType string               `json:"resourceType"`
	QuickID      string               `json:"quickId"`
	Labels       cmap.CustomStringMap `json:"labels"`
	Payload      string               `json:"payload"`
	PreDelay     string               `json:"preDelay"`
	Selector     string               `json:"selector"`
}

// WebhookData struct
type WebhookData struct {
	Server    string            `json:"server"`
	API       string            `json:"api"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Parameter string            `json:"parameter"`
	Body      interface{}       `json:"body"`
}

// EmailData struct
type EmailData struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}
