package node

import "time"

// Node model
type Node struct {
	ID        string                 `json:"id"`
	ShortID   string                 `json:"shortId"`
	GatewayID string                 `json:"gatewayId"`
	Name      string                 `json:"name"`
	ParentID  string                 `json:"parentId"`
	LastSeen  time.Time              `json:"lastSeen"`
	Config    map[string]interface{} `json:"config"`
	Others    map[string]interface{} `json:"others"`
	Labels    map[string]string      `json:"labels"`
}
