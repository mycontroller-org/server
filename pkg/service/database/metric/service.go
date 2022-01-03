package metrics

import (
	"errors"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	cfgTY "github.com/mycontroller-org/server/v2/pkg/types/config"
	metricPlugin "github.com/mycontroller-org/server/v2/plugin/database/metric"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	void_db "github.com/mycontroller-org/server/v2/plugin/database/metric/voiddb"
)

// Init metric database
func Init(metricCfg cmap.CustomMap, loggerCfg cfgTY.LoggerConfig) (metricTY.Plugin, error) {
	// include logger details
	metricCfg["logger"] = map[string]string{"mode": loggerCfg.Mode, "encoding": loggerCfg.Encoding, "level": loggerCfg.Level.Metric}

	if metricCfg.GetString(types.KeyType) == "" {
		return nil, errors.New("metric database type not defined")
	}

	updatedCfg := metricCfg.Clone()
	// if metric database disabled, supply void db
	if metricCfg.GetBool(types.KeyDisabled) {
		updatedCfg = cmap.CustomMap{}
		updatedCfg.Set(types.KeyType, void_db.PluginVoidDB, nil)
		updatedCfg.Set(types.KeyDisabled, false, nil)
	}

	plugin, err := metricPlugin.Create(updatedCfg.GetString(types.KeyType), updatedCfg)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}
