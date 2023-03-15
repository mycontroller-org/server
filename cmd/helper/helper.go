package helper

import (
	"context"
	"fmt"
	"time"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	bus "github.com/mycontroller-org/server/v2/pkg/bus"
	"github.com/mycontroller-org/server/v2/pkg/database"
	"github.com/mycontroller-org/server/v2/pkg/encryption"
	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	cfgTY "github.com/mycontroller-org/server/v2/pkg/types/config"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"github.com/mycontroller-org/server/v2/pkg/version"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// loads logger
func loadLogger(ctx context.Context, cfg cfgTY.LoggerConfig, component string) (context.Context, *zap.Logger) {
	logger := loggerUtils.GetLogger(cfg.Mode, cfg.Level.Core, cfg.Encoding, false, 0, cfg.EnableStacktrace)
	logger.Info(fmt.Sprintf("welcome to MyController %s :)", component))
	// do not print host Id on version information
	ver := version.Get()
	ver.HostID = ""
	logger.Info("version information", zap.Any("versionInfo", ver), zap.Any("loggerConfig", cfg))

	// in some places still using "z.L()...", which needs global logger should be enabled
	// enabling global logger.
	// to fix this, do `grep -rl "zap\.L()"` and fix those manually.
	zap.ReplaceGlobals(logger)

	return contextTY.LoggerWithContext(ctx, logger), logger
}

// set environment values, will be used by other sub components
func setEnvironmentVariables(cfg *cfgTY.Config) error {
	envMap := map[string]interface{}{
		types.ENV_DOCUMENTATION_URL:      cfg.Web.DocumentationURL,
		types.ENV_METRIC_DB_DISABLED:     cfg.Database.Metric.GetString(types.KeyDisabled),
		types.ENV_TELEMETRY_ENABLED:      cfg.Telemetry.Enabled,
		types.ENV_DIR_DATA:               cfg.Directories.GetData(),
		types.ENV_DIR_DATA_STORAGE:       cfg.Directories.GetDataStorage(),
		types.ENV_DIR_DATA_INTERNAL:      cfg.Directories.GetDataInternal(),
		types.ENV_DIR_DATA_FIRMWARE:      cfg.Directories.GetDataFirmware(),
		types.ENV_DIR_GATEWAY_LOGS:       cfg.Directories.GetGatewayLogs(),
		types.ENV_DIR_TMP:                cfg.Directories.GetTmp(),
		types.ENV_DIR_GATEWAY_TMP:        cfg.Directories.GetGatewayTmp(),
		types.ENV_LOG_LEVEL_CORE:         cfg.Logger.Level.Core,
		types.ENV_LOG_LEVEL_METRIC:       cfg.Logger.Level.Metric,
		types.ENV_LOG_LEVEL_STORAGE:      cfg.Logger.Level.Storage,
		types.ENV_LOG_LEVEL_WEB_HANDLER:  cfg.Logger.Level.WebHandler,
		types.ENV_LOG_MODE:               cfg.Logger.Mode,
		types.ENV_LOG_ENCODING:           cfg.Logger.Encoding,
		types.ENV_LOG_ENABLE_STACK_TRACE: cfg.Logger.EnableStacktrace,
		types.ENV_RUNNING_SINCE:          time.Now().Format(time.RFC3339),
	}

	for key, value := range envMap {
		if err := types.SetEnv(key, value); err != nil {
			return err
		}
	}
	return nil
}

// load encryption helper
func loadEncryptionHelper(ctx context.Context, logger *zap.Logger, secret string) (context.Context, *encryption.Encryption) {
	enc := encryption.New(logger, secret, nil, "")
	ctx = encryption.WithContext(ctx, enc)
	return ctx, enc
}

// load core scheduler
func loadCoreScheduler(ctx context.Context) (context.Context, schedulerTY.CoreScheduler) {
	coreScheduler := coreScheduler.New()
	ctx = schedulerTY.WithContext(ctx, coreScheduler)
	return ctx, coreScheduler
}

// load bus plugin
func loadBus(ctx context.Context, cfg cmap.CustomMap) (context.Context, busTY.Plugin, error) {
	_bus, err := bus.Get(ctx, cfg)
	if err != nil {
		return ctx, nil, err
	}
	ctx = busTY.WithContext(ctx, _bus)
	return ctx, _bus, nil
}

// load storage database
func loadStorageDatabase(ctx context.Context, cfg cmap.CustomMap) (context.Context, storageTY.Plugin, error) {
	storage, err := database.GetStorage(ctx, cfg)
	if err != nil {
		return ctx, nil, err
	}
	ctx = storageTY.WithContext(ctx, storage)
	return ctx, storage, nil
}

// load metric database
func loadMetricDatabase(ctx context.Context, cfg cmap.CustomMap) (context.Context, metricTY.Plugin, error) {
	metric, err := database.GetMetric(ctx, cfg)
	if err != nil {
		return ctx, nil, err
	}
	ctx = metricTY.WithContext(ctx, metric)
	return ctx, metric, nil
}

// load core api
func loadCoreApi(ctx context.Context) (context.Context, *entitiesAPI.API, error) {
	api, err := entitiesAPI.New(ctx)
	if err != nil {
		return ctx, nil, err
	}
	ctx = entitiesAPI.WithContext(ctx, api)
	return ctx, api, nil
}
