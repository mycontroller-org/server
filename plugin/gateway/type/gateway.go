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
	ID                 string               `json:"id"`
	Description        string               `json:"description"`
	Enabled            bool                 `json:"enabled"`
	ReconnectDelay     string               `json:"reconnectDelay"`
	QueueFailedMessage bool                 `json:"queueFailedMessage"`
	Provider           cmap.CustomMap       `json:"provider"`
	MessageLogger      cmap.CustomMap       `json:"messageLogger"`
	Labels             cmap.CustomStringMap `json:"labels"`
	Others             cmap.CustomMap       `json:"others"`
	State              *types.State         `json:"state"`
	ModifiedOn         time.Time            `json:"modifiedOn"`
	LastTransaction    time.Time            `json:"lastTransaction"`
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
