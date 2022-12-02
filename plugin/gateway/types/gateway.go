package gateway

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Gateway actions
const (
	ActionDiscoverNodes = "discover_nodes"
)

// Config struct
type Config struct {
	ID                 string               `json:"id" yaml:"id"`
	Description        string               `json:"description" yaml:"description"`
	Enabled            bool                 `json:"enabled" yaml:"enabled"`
	ReconnectDelay     string               `json:"reconnectDelay" yaml:"reconnectDelay"`
	QueueFailedMessage bool                 `json:"queueFailedMessage" yaml:"queueFailedMessage"`
	Provider           cmap.CustomMap       `json:"provider" yaml:"provider"`
	MessageLogger      cmap.CustomMap       `json:"messageLogger" yaml:"messageLogger"`
	Labels             cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Others             cmap.CustomMap       `json:"others" yaml:"others"`
	State              *types.State         `json:"state" yaml:"state"`
	ModifiedOn         time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
	LastTransaction    time.Time            `json:"lastTransaction" yaml:"lastTransaction"`
}

// GetReconnectDelay for this config
func (c *Config) GetReconnectDelay() *time.Duration {
	if c.ReconnectDelay == "" {
		return nil
	}

	duration, err := time.ParseDuration(c.ReconnectDelay)
	if err != nil {
		return nil
	}

	if duration < 1*time.Second {
		duration = time.Duration(1 * time.Second)
	}

	return &duration
}
