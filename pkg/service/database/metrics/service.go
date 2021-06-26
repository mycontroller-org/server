package metrics

import (
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	metricsML "github.com/mycontroller-org/server/v2/plugin/database/metrics"
	influx "github.com/mycontroller-org/server/v2/plugin/database/metrics/influxdb_v2"
	"github.com/mycontroller-org/server/v2/plugin/database/metrics/voiddb"
	"go.uber.org/zap"
)

// metrics database service
var (
	SVC      metricsML.Client
	Disabled bool
)

// Init metrics database
func Init(metricsCfg cmap.CustomMap) {
	// include logger details
	metricsCfg["logger"] = map[string]string{"mode": cfg.CFG.Logger.Mode, "encoding": cfg.CFG.Logger.Encoding, "level": cfg.CFG.Logger.Level.Metrics}

	dbType := metricsCfg.GetString("type")
	if dbType == "" {
		dbType = metricsML.TypeVoidDB
	} else if metricsCfg.GetBool("disabled") {
		dbType = metricsML.TypeVoidDB
	}
	switch dbType {
	case metricsML.TypeInfluxDB:
		Disabled = false
		client, err := influx.NewClient(metricsCfg)
		if err != nil {
			zap.L().Fatal("error on metrics database initialization", zap.Error(err))
		}
		SVC = client

	case metricsML.TypeVoidDB:
		Disabled = true
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
