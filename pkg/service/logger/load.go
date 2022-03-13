package logger

import (
	"fmt"

	cfgTY "github.com/mycontroller-org/server/v2/pkg/types/config"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"go.uber.org/zap"
)

// Load logger
func Load(loggerCfg cfgTY.LoggerConfig, component string) {
	logger := loggerUtils.GetLogger(loggerCfg.Mode, loggerCfg.Level.Core, loggerCfg.Encoding, false, 0, loggerCfg.EnableStacktrace)
	zap.ReplaceGlobals(logger)
	zap.L().Info(fmt.Sprintf("welcome to %s :)", component))
	//	zap.L().Info("server detail", zap.Any("version", version.Get()), zap.Any("loggerConfig", loggerCfg))
}
