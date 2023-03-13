package routes

import (
	"context"

	"github.com/gorilla/mux"
	action "github.com/mycontroller-org/server/v2/pkg/api/action"
	backupRestoreAPI "github.com/mycontroller-org/server/v2/pkg/api/backup"
	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	quickIdAPI "github.com/mycontroller-org/server/v2/pkg/api/quickid"
	export "github.com/mycontroller-org/server/v2/pkg/backup"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type Routes struct {
	ctx        context.Context
	logger     *zap.Logger
	storage    storageTY.Plugin
	enc        *encryptionAPI.Encryption
	metric     metricTY.Plugin
	api        *entitiesAPI.API
	action     *action.ActionAPI
	router     *mux.Router
	backupAPI  *backupRestoreAPI.BackupAPI
	quickIdAPI *quickIdAPI.QuickIdAPI
}

func New(ctx context.Context, router *mux.Router, enableProfiling bool) (*Routes, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	metric, err := metricTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entitiesAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	storage, err := storageTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	enc, err := encryptionAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	_action, err := action.New(ctx)
	if err != nil {
		return nil, err
	}
	_quickIdAPI, err := quickIdAPI.New(ctx)
	if err != nil {
		return nil, err
	}
	backupRestore, err := export.New(ctx)
	if err != nil {
		return nil, err
	}
	_backupAPI := backupRestoreAPI.New(ctx, logger, backupRestore, storage, bus, enc)

	routes := &Routes{
		ctx:        ctx,
		logger:     logger.Named("http_router"),
		storage:    storage,
		metric:     metric,
		api:        api,
		enc:        enc,
		action:     _action,
		router:     router,
		backupAPI:  _backupAPI,
		quickIdAPI: _quickIdAPI,
	}

	// register routes
	routes.registerActionRoutes()
	routes.registerBackupRestoreRoutes()
	routes.registerDashboardRoutes()
	routes.registerDataRepositoryRoutes()
	routes.registerFieldRoutes()
	routes.registerFirmwareRoutes()
	routes.registerForwardPayloadRoutes()
	routes.registerGatewayRoutes()
	routes.registerHandlerRoutes()
	routes.registerMetricRoutes()
	routes.registerNodeRoutes()
	routes.registerQuickIDRoutes()
	routes.registerSchedulerRoutes()
	routes.registerServiceTokenRoutes()
	routes.registerSourceRoutes()
	routes.registerStatusRoutes()
	routes.registerSystemRoutes()
	routes.registerTaskRoutes()
	routes.registerVirtualAssistantRoutes()
	routes.registerVirtualDeviceRoutes()

	// enables profiling
	if enableProfiling {
		routes.registerPProfRoutes()
	}

	return routes, nil
}
