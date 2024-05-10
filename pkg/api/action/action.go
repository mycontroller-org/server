package action

import (
	"context"

	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

type ActionAPI struct {
	logger *zap.Logger
	api    *entityAPI.API
	bus    busTY.Plugin
}

func New(ctx context.Context) (*ActionAPI, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &ActionAPI{
		logger: logger.Named("action_api"),
		api:    api,
		bus:    bus,
	}, nil
}
