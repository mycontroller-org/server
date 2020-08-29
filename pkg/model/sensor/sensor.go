package sensor

import (
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
)

// Sensor model
type Sensor struct {
	ID        string             `json:"id"`
	GatewayID string             `json:"gatewayId"`
	NodeID    string             `json:"nodeId"`
	Name      string             `json:"name"`
	Labels    ml.CustomStringMap `json:"labels"`
	Others    ml.CustomMap       `json:"others"`
	LastSeen  time.Time          `json:"lastSeen"`
}
