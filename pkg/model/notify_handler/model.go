package notifyhandler

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// operation types
const (
	TypeNoop           = "noop"
	TypeEmail          = "email"
	TypeTelegram       = "telegram"
	TypeWebhook        = "webhook"
	TypeSMS            = "sms"
	TypePushbullet     = "pushbullet"
	TypeResourceAction = "resource_action"
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
type MessageWrapper struct {
	ID        string
	Variables map[string]interface{}
}
