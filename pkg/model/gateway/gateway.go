package gateway

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
)

// Config struct
type Config struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Enabled  bool               `json:"enabled"`
	Ack      AckConfig          `json:"ack"`
	Provider ProviderConfig     `json:"provider"`
	Labels   ml.CustomStringMap `json:"labels"`
	Others   ml.CustomMap       `json:"others"`
	State    ml.State           `json:"state"`
}

// AckConfig data
type AckConfig struct {
	Enabled          bool   `json:"enabled"`
	StreamAckEnabled bool   `json:"streamAckEnabled"`
	RetryCount       int    `json:"retryCount"`
	Timeout          string `json:"timeout"`
}

// ProviderConfig data
type ProviderConfig struct {
	Type         string       `json:"type"`
	ProtocolType string       `json:"protocolType"`
	Config       ml.CustomMap `json:"config"`
}
