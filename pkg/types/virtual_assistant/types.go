package virtual_device

import (
	"net/http"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Config of virtual assistant
type Config struct {
	ID           string               `json:"id"`
	Description  string               `json:"description"`
	Enabled      bool                 `json:"enabled"`
	Labels       cmap.CustomStringMap `json:"labels"`
	DeviceFilter cmap.CustomStringMap `json:"deviceFilter"`
	ProviderType string               `json:"providerType"`
	Config       cmap.CustomMap       `json:"config"`
	State        *types.State         `json:"state"`
	ModifiedOn   time.Time            `json:"modifiedOn"`
}

// Virtual Assistant plugin interface
type Plugin interface {
	Start() error
	Stop() error
	Config() *Config
	ServeHTTP(http.ResponseWriter, *http.Request)
	Name() string
}
