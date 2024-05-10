package metric

import (
	"context"
	"errors"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	void_db "github.com/mycontroller-org/server/v2/plugin/database/metric/voiddb"
)

// returns metric database
func Get(ctx context.Context, metricCfg cmap.CustomMap) (metricTY.Plugin, error) {
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

	plugin, err := Create(ctx, updatedCfg.GetString(types.KeyType), updatedCfg)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}
