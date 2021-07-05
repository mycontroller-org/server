package storage

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	storageType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(config cmap.CustomMap) (storageType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(name string, config cmap.CustomMap) (p storageType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("storage database plugin [%s] is not registered", name)
	}
	return
}
