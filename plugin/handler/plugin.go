package handler

import (
	"fmt"

	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
)

// CreatorFn func type
type CreatorFn func(config *handlerType.Config) (handlerType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	creators[name] = fn
}

func Create(name string, config *handlerType.Config) (p handlerType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("handler plugin [%s] is not registered", name)
	}
	return
}
