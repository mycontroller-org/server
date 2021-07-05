package bus

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	busType "github.com/mycontroller-org/server/v2/plugin/bus/type"
)

// CreatorFn func type
type CreatorFn func(config cmap.CustomMap) (busType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	creators[name] = fn
}

func Create(name string, config cmap.CustomMap) (p busType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("bus plugin [%s] is not registered", name)
	}
	return
}
