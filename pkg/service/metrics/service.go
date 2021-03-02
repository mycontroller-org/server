package metrics

import (
	"errors"

	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	influx "github.com/mycontroller-org/backend/v2/plugin/metrics/influxdb_v2"
	"github.com/mycontroller-org/backend/v2/plugin/metrics/voiddb"
	"go.uber.org/zap"
)

// metrics database service
var (
	SVC mtsml.Client
)

// Init metrics database
func Init() {
	metricsCfg, err := getDatabaseConfig(cfg.CFG.Database.Metrics)
	if err != nil {
		zap.L().Fatal("Problem with metrics database config", zap.String("name", cfg.CFG.Database.Metrics), zap.Error(err))
	}

	// include logger details
	metricsCfg["logger"] = map[string]string{"mode": cfg.CFG.Logger.Mode, "encoding": cfg.CFG.Logger.Encoding, "level": cfg.CFG.Logger.Level.Metrics}

	dbType, available := metricsCfg["type"]
	if available {
		switch dbType {
		case mtsml.TypeInfluxdbV2:
			client, err := influx.NewClient(metricsCfg)
			if err != nil {
				zap.L().Fatal("error on metrics database initialization", zap.Error(err), zap.String("database", cfg.CFG.Database.Metrics))
			}
			SVC = client

		case mtsml.TypeVoidDB:
			client, err := voiddb.NewClient(metricsCfg)
			if err != nil {
				zap.L().Fatal("error on metrics database initialization", zap.Error(err), zap.String("database", cfg.CFG.Database.Metrics))
			}
			SVC = client

		default:
			zap.L().Fatal("Specified database type not implemented", zap.Any("type", dbType), zap.String("database", cfg.CFG.Database.Metrics))
		}
		return
	}
	zap.L().Fatal("'type' field should be added on the database config", zap.String("database", cfg.CFG.Database.Metrics))
}

func getDatabaseConfig(name string) (map[string]interface{}, error) {
	for _, d := range cfg.CFG.Databases {
		if d["name"] == name {
			return d, nil
		}
	}
	return nil, errors.New("Config not found")
}
