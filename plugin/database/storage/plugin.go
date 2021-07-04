package storage

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	storageType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// CreatorFn func type
type CreatorFn func(config cmap.CustomMap) (storageType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	creators[name] = fn
}

func Create(name string, config cmap.CustomMap) (p storageType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("storage plugin [%s] is not registered", name)
	}
	return
}
