package notifyhandler

import (
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// handler types
const (
	TypeEmail = "email"
)

// Config model
type Config struct {
	ID     string               `json:"id"`
	Type   string               `json:"type"`
	Labels cmap.CustomStringMap `json:"labels"`
	Spec   cmap.CustomMap       `json:"spec"`
}

// Clone config
func (svc *Config) Clone() Config {
	newServices := Config{
		ID:     svc.ID,
		Type:   svc.Type,
		Labels: svc.Labels.Clone(),
		Spec:   svc.Spec.Clone(),
	}
	return newServices
}
