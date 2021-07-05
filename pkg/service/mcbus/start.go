package mcbus

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	busPlugin "github.com/mycontroller-org/server/v2/plugin/bus"
	busType "github.com/mycontroller-org/server/v2/plugin/bus/type"
	"go.uber.org/zap"
)

var (
	busClient busType.Plugin
	pauseSRV  concurrency.SafeBool
)

// Start function
func Start(config cmap.CustomMap) {
	// get plugin type
	pluginType := config.GetString(model.KeyType)
	if pluginType == "" {
		zap.L().Fatal("bus plugin type is not defined")
		return
	}

	// update topic prefix
	topicPrefix = config.GetString(keyTopicPrefix)

	// update client
	plugin, err := busPlugin.Create(pluginType, config)
	if err != nil {
		zap.L().Fatal("failed to get bus client", zap.Error(err))
	}

	pauseSRV = concurrency.SafeBool{}
	busClient = plugin
}
