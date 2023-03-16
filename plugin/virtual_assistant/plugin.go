package virtual_assistant

import (
	"context"
	"fmt"

	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(ctx context.Context, config *vaTY.Config) (vaTY.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(ctx context.Context, name string, config *vaTY.Config) (p vaTY.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(ctx, config)
	} else {
		err = fmt.Errorf("virtual assistant plugin [%s] is not registered", name)
	}
	return
}
