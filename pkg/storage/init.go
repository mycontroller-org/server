package storage

import (
	"errors"
	"fmt"

	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	sch "github.com/mycontroller-org/backend/v2/pkg/scheduler"
	"github.com/mycontroller-org/backend/v2/plugin/storage/memory"
	"github.com/mycontroller-org/backend/v2/plugin/storage/mongodb"
)

// Init storage
func Init(config map[string]interface{}, scheduler *sch.Scheduler) (stgml.Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		case stgml.DBTypeMemory:
			return memory.NewClient(config, scheduler)
		case stgml.DBTypeMongoDB:
			return mongodb.NewClient(config)
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}
