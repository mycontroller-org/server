package metric

import (
	"context"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(ctx context.Context, config cmap.CustomMap) (metricTY.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(ctx context.Context, name string, config cmap.CustomMap) (p metricTY.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(ctx, config)
	} else {
		err = fmt.Errorf("metric database plugin [%s] is not registered", name)
	}
	return
}
