package entity_api

import (
	"context"
	"errors"

	dashboard "github.com/mycontroller-org/server/v2/pkg/api/dashboard"
	dataRepository "github.com/mycontroller-org/server/v2/pkg/api/data_repository"
	field "github.com/mycontroller-org/server/v2/pkg/api/field"
	firmware "github.com/mycontroller-org/server/v2/pkg/api/firmware"
	forwardPayload "github.com/mycontroller-org/server/v2/pkg/api/forward_payload"
	gateway "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	handler "github.com/mycontroller-org/server/v2/pkg/api/handler"
	node "github.com/mycontroller-org/server/v2/pkg/api/node"
	schedule "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	serviceToken "github.com/mycontroller-org/server/v2/pkg/api/service_token"
	settings "github.com/mycontroller-org/server/v2/pkg/api/settings"
	source "github.com/mycontroller-org/server/v2/pkg/api/source"
	status "github.com/mycontroller-org/server/v2/pkg/api/status"
	sunrise "github.com/mycontroller-org/server/v2/pkg/api/sunrise"
	task "github.com/mycontroller-org/server/v2/pkg/api/task"
	user "github.com/mycontroller-org/server/v2/pkg/api/user"
	virtualAssistant "github.com/mycontroller-org/server/v2/pkg/api/virtual_assistant"
	virtualDevice "github.com/mycontroller-org/server/v2/pkg/api/virtual_device"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	"github.com/mycontroller-org/server/v2/pkg/types"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const (
	contextKey types.ContextKey = "entity_api"
)

type API struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
	enc     *encryptionAPI.Encryption
}

func FromContext(ctx context.Context) (*API, error) {
	api, ok := ctx.Value(contextKey).(*API)
	if !ok {
		return nil, errors.New("invalid entity api instance received in context")
	}
	if api == nil {
		return nil, errors.New("entity api instance not provided in context")
	}
	return api, nil
}

func WithContext(ctx context.Context, api *API) context.Context {
	return context.WithValue(ctx, contextKey, api)
}

func New(ctx context.Context) (*API, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	storage, err := storageTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	enc, err := encryptionAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &API{
		ctx:     ctx,
		logger:  logger.Named("api"),
		storage: storage,
		bus:     bus,
		enc:     enc,
	}, nil
}

func (a *API) Dashboard() *dashboard.DashboardAPI {
	return dashboard.New(a.ctx, a.logger, a.storage)
}

func (a *API) DataRepository() *dataRepository.DataRepositoryAPI {
	return dataRepository.New(a.ctx, a.logger, a.storage, a.enc, a.bus)
}

func (a *API) Field() *field.FieldAPI {
	return field.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) Firmware() *firmware.FirmwareAPI {
	return firmware.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) ForwardPayload() *forwardPayload.ForwardPayloadAPI {
	return forwardPayload.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) Gateway() *gateway.GatewayAPI {
	return gateway.New(a.ctx, a.logger, a.storage, a.enc, a.bus)
}

func (a *API) Handler() *handler.HandlerAPI {
	return handler.New(a.ctx, a.logger, a.storage, a.enc, a.bus)
}

func (a *API) Node() *node.NodeAPI {
	return node.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) Schedule() *schedule.ScheduleAPI {
	return schedule.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) ServiceToken() *serviceToken.ServiceTokenAPI {
	return serviceToken.New(a.ctx, a.logger, a.storage)
}
func (a *API) Settings() *settings.SettingsAPI {
	return settings.New(a.ctx, a.logger, a.storage, a.enc, a.bus)
}

func (a *API) Source() *source.SourceAPI {
	return source.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) Status() *status.StatusAPI {
	return status.New(a.ctx, a.logger, a.storage, a.enc, a.bus)
}

func (a *API) Sunrise() types.Sunrise {
	return sunrise.New(a.ctx, a.logger, a.storage, a.enc, a.bus)
}

func (a *API) Task() *task.TaskAPI {
	return task.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) User() *user.UserAPI {
	return user.New(a.ctx, a.logger, a.storage)
}

func (a *API) VirtualAssistant() *virtualAssistant.VirtualAssistantAPI {
	return virtualAssistant.New(a.ctx, a.logger, a.storage, a.bus)
}

func (a *API) VirtualDevice() *virtualDevice.VirtualDeviceAPI {
	return virtualDevice.New(a.ctx, a.logger, a.storage, a.bus)
}
