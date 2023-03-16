package virtual_assistant

import (
	"net/http"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Config of virtual assistant
type Config struct {
	ID           string               `json:"id" yaml:"id"`
	Description  string               `json:"description" yaml:"description"`
	Enabled      bool                 `json:"enabled" yaml:"enabled"`
	Labels       cmap.CustomStringMap `json:"labels" yaml:"labels"`
	DeviceFilter cmap.CustomStringMap `json:"deviceFilter" yaml:"deviceFilter"`
	ProviderType string               `json:"providerType" yaml:"providerType"`
	Config       cmap.CustomMap       `json:"config" yaml:"config"`
	State        *types.State         `json:"state" yaml:"state"`
	ModifiedOn   time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}

// Virtual Assistant plugin interface
type Plugin interface {
	Start() error
	Stop() error
	Config() *Config
	ServeHTTP(http.ResponseWriter, *http.Request)
	Name() string
}
