package metrics

import (
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	cfgML "github.com/mycontroller-org/server/v2/pkg/model/config"
	metricPlugin "github.com/mycontroller-org/server/v2/plugin/database/metric"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	void_db "github.com/mycontroller-org/server/v2/plugin/database/metric/voiddb"
)

// Init metric database
func Init(metricCfg cmap.CustomMap, loggerCfg cfgML.LoggerConfig) (metricType.Plugin, error) {
	// include logger details
	metricCfg["logger"] = map[string]string{"mode": loggerCfg.Mode, "encoding": loggerCfg.Encoding, "level": loggerCfg.Level.Metric}

	if metricCfg.GetString(model.KeyType) == "" {
		return nil, errors.New("metric database type not defined")
	}

	updatedCfg := metricCfg.Clone()
	// if metric database disabled, supply void db
	if metricCfg.GetBool(model.KeyDisabled) {
		updatedCfg = cmap.CustomMap{}
		updatedCfg.Set(model.KeyType, void_db.PluginVoidDB, nil)
		updatedCfg.Set(model.KeyDisabled, false, nil)
	}

	plugin, err := metricPlugin.Create(updatedCfg.GetString(model.KeyType), updatedCfg)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}
