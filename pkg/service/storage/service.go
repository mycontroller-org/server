package storage

import (
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
	"github.com/mycontroller-org/server/v2/plugin/database/storage/memory"
	"github.com/mycontroller-org/server/v2/plugin/database/storage/mongodb"
	"go.uber.org/zap"
)

// storage service
var (
	SVC stgML.Client
)

// Init storage service
func Init(storageCfg map[string]interface{}, importFunc func(targetDir, fileType string, ignoreEmptyDir bool) error) {
	// include logger details
	storageCfg["logger"] = map[string]string{"mode": cfg.CFG.Logger.Mode, "encoding": cfg.CFG.Logger.Encoding, "level": cfg.CFG.Logger.Level.Storage}

	// init storage database
	dbType, available := storageCfg["type"]
	if available {
		switch dbType {
		case stgML.TypeMemory:
			client, err := memory.NewClient(storageCfg)
			if err != nil {
				zap.L().Fatal("error on storage database initialization", zap.Error(err))
			}
			SVC = client
			// run local import
			err = client.LocalImport(importFunc)
			if err != nil {
				zap.L().Fatal("error on run local import on memory database", zap.Error(err))
			}

		case stgML.TypeMongoDB:
			client, err := mongodb.NewClient(storageCfg)
			if err != nil {
				zap.L().Fatal("error on storage database initialization", zap.Error(err))
			}
			SVC = client

		default:
			zap.L().Fatal("specified database type not implemented", zap.Any("type", dbType))
		}
		return
	}
	zap.L().Fatal("'type' field should be added on the database config")

}
