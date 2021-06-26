package metrics

import (
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	mtsML "github.com/mycontroller-org/server/v2/plugin/database/metrics"
	influx "github.com/mycontroller-org/server/v2/plugin/database/metrics/influxdb_v2"
	"github.com/mycontroller-org/server/v2/plugin/database/metrics/voiddb"
	"go.uber.org/zap"
)

// metrics database service
var (
	SVC mtsML.Client
)

// Init metrics database
func Init(metricsCfg map[string]interface{}) {
	// include logger details
	metricsCfg["logger"] = map[string]string{"mode": cfg.CFG.Logger.Mode, "encoding": cfg.CFG.Logger.Encoding, "level": cfg.CFG.Logger.Level.Metrics}

	dbType, available := metricsCfg["type"]
	if available {
		switch dbType {
		case mtsML.TypeInfluxDB:
			client, err := influx.NewClient(metricsCfg)
			if err != nil {
				zap.L().Fatal("error on metrics database initialization", zap.Error(err))
			}
			SVC = client

		case mtsML.TypeVoidDB:
			client, err := voiddb.NewClient(metricsCfg)
			if err != nil {
				zap.L().Fatal("error on metrics database initialization", zap.Error(err))
			}
			SVC = client

		default:
			zap.L().Fatal("specified database type not implemented", zap.Any("type", dbType))
		}
		return
	}
	zap.L().Fatal("'type' field should be added on the database config")
}
