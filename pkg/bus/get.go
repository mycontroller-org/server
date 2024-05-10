package mcbus

import (
	"context"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busPlugin "github.com/mycontroller-org/server/v2/plugin/bus"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// return bus instance
func Get(ctx context.Context, config cmap.CustomMap) (busTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	// get plugin type
	pluginType := config.GetString(types.KeyType)
	if pluginType == "" {
		logger.Fatal("bus plugin type is not defined")
		return nil, err
	}

	// update client
	bus, err := busPlugin.Create(ctx, pluginType, config)
	if err != nil {
		logger.Fatal("failed to get bus client", zap.Error(err))
		return nil, err
	}

	return bus, nil
}
