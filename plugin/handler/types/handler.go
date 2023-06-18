package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// handler data types
const (
	DataTypeResource = "resource"
	DataTypeEmail    = "email"
	DataTypeTelegram = "telegram"
	DataTypeWebhook  = "webhook"
	DataTypeBackup   = "backup"
)

var (
	ErrReQueue = errors.New("requeue")
)

// Plugin interface details for operation
type Plugin interface {
	Name() string
	Start() error
	Close() error
	Post(parameters map[string]interface{}) error
	State() *types.State
}

// Config struct
type Config struct {
	ID          string               `json:"id" yaml:"id"`
	Description string               `json:"description" yaml:"description"`
	Enabled     bool                 `json:"enabled" yaml:"enabled"`
	Labels      cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Type        string               `json:"type" yaml:"type"`
	Spec        cmap.CustomMap       `json:"spec" yaml:"spec"`
	ModifiedOn  time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
	State       *types.State         `json:"state" yaml:"state"`
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

// used in bus
// specially used to send data to handlers
type MessageWrapper struct {
	ID   string
	Data map[string]interface{}
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

// simple string data
type StringData struct {
	Disabled string `json:"disabled" yaml:"disabled"`
	Type     string `json:"type" yaml:"type"`
	Value    string `json:"value" yaml:"value"`
}

// ResourceData struct
type ResourceData struct {
	Disabled     string               `json:"disabled" yaml:"disabled"`
	Type         string               `json:"type" yaml:"type"`
	ResourceType string               `json:"resourceType" yaml:"resourceType"`
	QuickID      string               `json:"quickId" yaml:"quickId"`
	Labels       cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Payload      string               `json:"payload" yaml:"payload"`
	PreDelay     string               `json:"preDelay" yaml:"preDelay"`
	KeyPath      string               `json:"keyPath" yaml:"keyPath"`
}

// WebhookData struct
type WebhookData struct {
	Disabled        string                 `json:"disabled" yaml:"disabled"`
	Type            string                 `json:"type" yaml:"type"`
	Server          string                 `json:"server" yaml:"server"`
	API             string                 `json:"api" yaml:"api"`
	Insecure        bool                   `json:"insecure" yaml:"insecure"`
	Method          string                 `json:"method" yaml:"method"`
	Headers         map[string]string      `json:"headers" yaml:"headers"`
	QueryParameters map[string]interface{} `json:"queryParameters" yaml:"queryParameters"`
	Data            interface{}            `json:"data" yaml:"data"`
	CustomData      bool                   `json:"customData" yaml:"customData"`
	ResponseCode    int                    `json:"responseCode" yaml:"responseCode"`
}

// EmailData struct
type EmailData struct {
	Disabled string   `json:"disabled" yaml:"disabled"`
	Type     string   `json:"type" yaml:"type"`
	From     string   `json:"from" yaml:"from"`
	To       []string `json:"to" yaml:"to"`
	Subject  string   `json:"subject" yaml:"subject"`
	Body     string   `json:"body" yaml:"body"`
}

// TelegramData struct
type TelegramData struct {
	Disabled  string   `json:"disabled" yaml:"disabled"`
	Type      string   `json:"type" yaml:"type"`
	ChatIDs   []string `json:"chatIds" yaml:"chatIds"`
	ParseMode string   `json:"parseMode" yaml:"parseMode"`
	Text      string   `json:"text" yaml:"text"`
}

// BackupData struct
type BackupData struct {
	Type         string `json:"type" yaml:"type"`
	ProviderType string `json:"providerType" yaml:"providerType"`
	// additional fields will be added and used in the specific section
}

func IsTypeOf(parameter interface{}, expectedType string) (cmap.CustomMap, bool) {
	if _parameterMap, ok := parameter.(map[string]interface{}); ok {
		parameterCMap := cmap.CustomMap(_parameterMap)
		return parameterCMap, parameterCMap.GetString(types.KeyType) == expectedType
	}
	return nil, false
}

func HasTypePrefixOf(parameter interface{}, expectedType string) (cmap.CustomMap, bool) {
	if _parameterMap, ok := parameter.(map[string]interface{}); ok {
		parameterCMap := cmap.CustomMap(_parameterMap)
		return parameterCMap, strings.HasPrefix(parameterCMap.GetString(types.KeyType), expectedType)
	}
	return nil, false
}
