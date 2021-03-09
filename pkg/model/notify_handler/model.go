package notifyhandler

import (
	"time"

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
	TypeExporter   = "exporter"
)

// handler data types
const (
	DataTypeEmail      = "email"
	DataTypeTelegram   = "telegram"
	DataTypeWebhook    = "webhook"
	DataTypeSMS        = "sms"
	DataTypePushbullet = "pushbullet"
	DataTypeResource   = "resource"
	DataTypeExporter   = "exporter"
)

// Config model
type Config struct {
	ID          string               `json:"id"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	Labels      cmap.CustomStringMap `json:"labels"`
	Type        string               `json:"type"`
	Spec        cmap.CustomMap       `json:"spec"`
	ModifiedOn  time.Time            `json:"modifiedOn"`
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

// GenericData struct
type GenericData struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// ResourceData struct
type ResourceData struct {
	ResourceType string               `yaml:"resourceType"`
	QuickID      string               `yaml:"quickId"`
	Labels       cmap.CustomStringMap `yaml:"labels"`
	Payload      string               `yaml:"payload"`
	PreDelay     string               `yaml:"preDelay"`
	Selector     string               `yaml:"selector"`
}

// WebhookData struct
type WebhookData struct {
	Server    string            `yaml:"server"`
	API       string            `yaml:"api"`
	Method    string            `yaml:"method"`
	Headers   map[string]string `yaml:"headers"`
	Parameter string            `yaml:"parameter"`
	Body      interface{}       `yaml:"body"`
}

// EmailData struct
type EmailData struct {
	From    string   `yaml:"from"`
	To      []string `yaml:"to"`
	Subject string   `yaml:"subject"`
	Body    string   `yaml:"body"`
}

// TelegramData struct
type TelegramData struct {
	ChatIDs   []string `yaml:"chatIds"`
	ParseMode string   `yaml:"parseMode"`
	Text      string   `yaml:"text"`
}

// ExporterData struct
type ExporterData struct {
	ExporterType string                 `yaml:"exporterType"`
	Spec         map[string]interface{} `yaml:"spec"`
}
