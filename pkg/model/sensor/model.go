package sensor

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Sensor model
type Sensor struct {
	ID        string               `json:"id"`
	GatewayID string               `json:"gatewayId"`
	NodeID    string               `json:"nodeId"`
	Name      string               `json:"name"`
	Labels    cmap.CustomStringMap `json:"labels"`
	Others    cmap.CustomMap       `json:"others"`
	LastSeen  time.Time            `json:"lastSeen"`
}

// AssembleID forms actual ID using the supplied gatewayID, nodeID and sensorID
func AssembleID(gatewayID, nodeID, sensorID string) string {
	return fmt.Sprintf("%s_%s_%s", gatewayID, nodeID, sensorID)
}
