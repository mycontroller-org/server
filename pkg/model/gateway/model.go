package gateway

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Config struct
type Config struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	Enabled  bool                 `json:"enabled"`
	Ack      AckConfig            `json:"ack"`
	Provider ProviderConfig       `json:"provider"`
	Labels   cmap.CustomStringMap `json:"labels"`
	Others   cmap.CustomMap       `json:"others"`
	State    ml.State             `json:"state"`
}

// AckConfig data
type AckConfig struct {
	Enabled    bool   `json:"enabled"`
	RetryCount int    `json:"retryCount"`
	Timeout    string `json:"timeout"`
}

// ProviderConfig data
type ProviderConfig struct {
	Type         string         `json:"type"`
	ProtocolType string         `json:"protocolType"`
	Config       cmap.CustomMap `json:"config"`
}
