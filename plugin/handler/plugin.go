package handler

import (
	"context"
	"fmt"

	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// CreatorFn func type
type CreatorFn func(ctx context.Context, config *handlerTY.Config) (handlerTY.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	if _, found := creators[name]; found {
		zap.L().Fatal("duplicate plugin found", zap.String("pluginName", name))
		return
	}
	creators[name] = fn
}

func Create(ctx context.Context, name string, config *handlerTY.Config) (p handlerTY.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(ctx, config)
	} else {
		err = fmt.Errorf("handler plugin [%s] is not registered", name)
	}
	return
}
