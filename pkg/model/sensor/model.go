package sensor

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Sensor model
type Sensor struct {
	ID         string               `json:"id"`
	GatewayID  string               `json:"gatewayId"`
	NodeID     string               `json:"nodeId"`
	SensorID   string               `json:"sensorId"`
	Name       string               `json:"name"`
	Labels     cmap.CustomStringMap `json:"labels"`
	Others     cmap.CustomMap       `json:"others"`
	LastSeen   time.Time            `json:"lastSeen"`
	ModifiedOn time.Time            `json:"modifiedOn"`
}
