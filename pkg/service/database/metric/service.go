package metrics

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	metricPlugin "github.com/mycontroller-org/server/v2/plugin/database/metric"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	"go.uber.org/zap"
)

// metrics database service
var (
	SVC      metricType.Plugin
	Disabled bool
)

// Init metric database
func Init(metricCfg cmap.CustomMap) {
	// include logger details
	metricCfg["logger"] = map[string]string{"mode": cfg.CFG.Logger.Mode, "encoding": cfg.CFG.Logger.Encoding, "level": cfg.CFG.Logger.Level.Metric}

	pluginType := metricCfg.GetString(model.KeyType)

	if pluginType == "" {
		zap.L().Fatal("metric database type not defined")
		return
	}

	plugin, err := metricPlugin.Create(pluginType, metricCfg)
	if err != nil {
		zap.L().Fatal("error on metric database initialization", zap.Error(err))
		return
	}

	SVC = plugin
}
