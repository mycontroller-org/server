package storage

import (
	"fmt"

	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	vaAlexa "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/alexa"
	vaGoogle "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/google"
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

// init plugins
func init() {
	Register(vaGoogle.PluginGoogleAssistant, vaGoogle.New)
	Register(vaAlexa.PluginAlexaAssistant, vaAlexa.New)
}
