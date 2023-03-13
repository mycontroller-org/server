package gateway

import (
	"context"
	"fmt"

	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwPluginTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(ctx context.Context, config *gwPluginTY.Config) (providerTY.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

// logger *zap.Logger, gatewayCfg *gateway.Config, scheduler schedule.CoreScheduler
func Create(ctx context.Context, name string, config *gwPluginTY.Config) (p providerTY.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(ctx, config)
	} else {
		err = fmt.Errorf("gateway plugin [%s] is not registered", name)
	}
	return
}
