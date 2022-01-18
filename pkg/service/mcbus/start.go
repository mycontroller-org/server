package mcbus

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	busPlugin "github.com/mycontroller-org/server/v2/plugin/bus"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

var (
	busClient busTY.Plugin
	pauseSRV  concurrency.SafeBool
)

// Start function
func Start(config cmap.CustomMap) {
	// get plugin type
	pluginType := config.GetString(types.KeyType)
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
