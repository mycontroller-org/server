package metrics

import (
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	cfgML "github.com/mycontroller-org/server/v2/pkg/model/config"
	metricPlugin "github.com/mycontroller-org/server/v2/plugin/database/metric"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
)

// Init metric database
func Init(metricCfg cmap.CustomMap, loggerCfg cfgML.LoggerConfig) (metricType.Plugin, error) {
	// include logger details
	metricCfg["logger"] = map[string]string{"mode": loggerCfg.Mode, "encoding": loggerCfg.Encoding, "level": loggerCfg.Level.Metric}

	pluginType := metricCfg.GetString(model.KeyType)

	if pluginType == "" {
		return nil, errors.New("metric database type not defined")
	}

	plugin, err := metricPlugin.Create(pluginType, metricCfg)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}
