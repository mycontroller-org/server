package virtual_assistant

import (
	"fmt"

	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(config *vaTY.Config) (vaTY.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(name string, config *vaTY.Config) (p vaTY.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("virtual assistant plugin [%s] is not registered", name)
	}
	return
}
