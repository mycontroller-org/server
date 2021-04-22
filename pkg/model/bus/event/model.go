package event

// event types
const (
	TypeCreated   = "created"
	TypeUpdated   = "updated"
	TypeDeleted   = "deleted"
	TypeRequested = "requested"
)

// Event struct
type Event struct {
	Type          string      `json:"type"`
	EntityType    string      `json:"entityType"`
	EntityID      string      `json:"entityId"`
	EntityQuickID string      `json:"entityQuickId"`
	Entity        interface{} `json:"entity"`
}
