package utils

import (
	"context"
	"fmt"

	cfgTY "github.com/mycontroller-org/server/v2/pkg/types/config"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	"go.uber.org/zap"
)

// Load logger
func Load(ctx context.Context, loggerCfg cfgTY.LoggerConfig, component string) context.Context {
	logger := GetLogger(loggerCfg.Mode, loggerCfg.Level.Core, loggerCfg.Encoding, false, 0, loggerCfg.EnableStacktrace)
	zap.L().Info(fmt.Sprintf("welcome to %s :)", component))
	//	zap.L().Info("server detail", zap.Any("version", version.Get()), zap.Any("loggerConfig", loggerCfg))
	return contextTY.LoggerWithContext(ctx, logger)
}
