package service

import (
	"errors"
	"fmt"

	sch "github.com/mycontroller-org/backend/v2/pkg/scheduler"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"github.com/mycontroller-org/backend/v2/plugin/storage/memory"
	"github.com/mycontroller-org/backend/v2/plugin/storage/mongodb"
)

// InitStorageDatabase storage
func InitStorageDatabase(config map[string]interface{}, scheduler *sch.Scheduler) (stgml.Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		case stgml.TypeMemory:
			return memory.NewClient(config, scheduler)
		case stgml.TypeMongoDB:
			return mongodb.NewClient(config)
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}
