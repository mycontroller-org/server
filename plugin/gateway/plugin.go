package gateway

import (
	"fmt"

	providerType "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwType "github.com/mycontroller-org/server/v2/plugin/gateway/type"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(config *gwType.Config) (providerType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(name string, config *gwType.Config) (p providerType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("gateway plugin [%s] is not registered", name)
	}
	return
}
