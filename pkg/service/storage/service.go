package storage

import (
	"errors"

	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"github.com/mycontroller-org/backend/v2/plugin/storage/memory"
	"github.com/mycontroller-org/backend/v2/plugin/storage/mongodb"
	"go.uber.org/zap"
)

// storage service
var (
	SVC stgml.Client
)

// Init storage service
func Init() {
	// Get storage and metric database config
	storageCfg, err := getDatabaseConfig(cfg.CFG.Database.Storage)
	if err != nil {
		zap.L().Fatal("Problem with storage database config", zap.String("name", cfg.CFG.Database.Storage), zap.Error(err))
	}

	// include logger details
	storageCfg["logger"] = map[string]string{"mode": cfg.CFG.Logger.Mode, "encoding": cfg.CFG.Logger.Encoding, "level": cfg.CFG.Logger.Level.Storage}

	// init storage database
	dbType, available := storageCfg["type"]
	if available {
		switch dbType {
		case stgml.TypeMemory:
			client, err := memory.NewClient(storageCfg)
			if err != nil {
				zap.L().Fatal("error on storage database initialization", zap.Error(err), zap.String("database", cfg.CFG.Database.Storage))
			}
			SVC = client
		case stgml.TypeMongoDB:
			client, err := mongodb.NewClient(storageCfg)
			if err != nil {
				zap.L().Fatal("error on storage database initialization", zap.Error(err), zap.String("database", cfg.CFG.Database.Storage))
			}
			SVC = client
		default:
			zap.L().Fatal("Specified database type not implemented", zap.Any("type", dbType), zap.String("database", cfg.CFG.Database.Storage))
		}
		return
	}
	zap.L().Fatal("'type' field should be added on the database config", zap.String("database", cfg.CFG.Database.Storage))

}

func getDatabaseConfig(name string) (map[string]interface{}, error) {
	for _, d := range cfg.CFG.Databases {
		if d["name"] == name {
			return d, nil
		}
	}
	return nil, errors.New("Config not found")
}
