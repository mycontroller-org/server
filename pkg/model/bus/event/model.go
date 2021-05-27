package event

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/utils"
)

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

func (e *Event) LoadEntity(out interface{}) error {
	mapData, ok := e.Entity.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data type: %T", e.Entity)
	}
	return utils.MapToStruct(utils.TagNameJSON, mapData, out)
}
