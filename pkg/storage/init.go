package storage

import (
	"errors"
	"fmt"

	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	"github.com/mycontroller-org/backend/v2/pkg/scheduler"
	"github.com/mycontroller-org/backend/v2/plugin/storage/memory"
	"github.com/mycontroller-org/backend/v2/plugin/storage/mongodb"
)

// Init storage
func Init(config map[string]interface{}, sch *scheduler.Scheduler) (stgml.Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		/*
			// NOTE: badger support needs gcc installed on the build environment
			// addeds additional 6 MB on application bin file
				case TypeBadger:
					c, err := badger.NewClient(config)
					if err != nil {
						return nil, err
					}
					var cl Client = c
					return &cl, nil
		*/
		case stgml.TypeMemory:
			return memory.NewClient(config, sch)
		case stgml.TypeMongoDB:
			return mongodb.NewClient(config)
		//case stgml.TypeSqlite:
		//	return sqlite.NewClient(config)
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}
