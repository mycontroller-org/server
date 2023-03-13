package event

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/utils"
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
	Type          string      `json:"type" yaml:"type"`
	EntityType    string      `json:"entityType" yaml:"entityType"`
	EntityID      string      `json:"entityId" yaml:"entityId"`
	EntityQuickID string      `json:"entityQuickId" yaml:"entityQuickId"`
	Entity        interface{} `json:"entity" yaml:"entity"`
}

func (e *Event) LoadEntity(out interface{}) error {
	mapData, ok := e.Entity.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data type: %T", e.Entity)
	}
	return utils.MapToStruct(utils.TagNameJSON, mapData, out)
}
