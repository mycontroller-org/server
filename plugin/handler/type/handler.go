package handler

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
)

// handler data types
const (
	DataTypeResource = "resource"
	DataTypeEmail    = "email"
	DataTypeTelegram = "telegram"
	DataTypeWebhook  = "webhook"
	DataTypeBackup   = "backup"
)

// Plugin interface details for operation
type Plugin interface {
	Name() string
	Start() error
	Close() error
	Post(variables map[string]interface{}) error
	State() *model.State
}

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
	Type     string `json:"type"`
	Disabled string `json:"disabled"` // string? supports template
	Data     string `json:"data"`
}

// // ConvertibleBoolean used to convert string to bool
// type ConvertibleBoolean bool
//
// func (cb *ConvertibleBoolean) UnmarshalJSON(data []byte) error {
// 	asString := string(data)
// 	if converterUtils.ToBool(asString) {
// 		*cb = true
// 	} else {
// 		*cb = false
// 	}
// 	return nil
// }

// ResourceData struct
type ResourceData struct {
	ResourceType string               `yaml:"resourceType"`
	QuickID      string               `yaml:"quickId"`
	Labels       cmap.CustomStringMap `yaml:"labels"`
	Payload      string               `yaml:"payload"`
	PreDelay     string               `yaml:"preDelay"`
	KeyPath      string               `yaml:"keyPath"`
}

// WebhookData struct
type WebhookData struct {
	Server             string                 `yaml:"server"`
	API                string                 `yaml:"api"`
	InsecureSkipVerify bool                   `json:"insecureSkipVerify"`
	Method             string                 `yaml:"method"`
	Headers            map[string]string      `yaml:"headers"`
	QueryParameters    map[string]interface{} `yaml:"queryParameters"`
	Data               interface{}            `yaml:"data"`
	CustomData         bool                   `yaml:"customData"`
	ResponseCode       int                    `yaml:"responseCode"`
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

// BackupData struct
type BackupData struct {
	ProviderType string                 `yaml:"providerType"`
	Spec         map[string]interface{} `yaml:"spec"`
}
