package bus

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	busPluginTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(config cmap.CustomMap) (busPluginTY.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(name string, config cmap.CustomMap) (p busPluginTY.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("bus plugin [%s] is not registered", name)
	}
	return
}
