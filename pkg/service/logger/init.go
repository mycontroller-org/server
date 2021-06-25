package logger

import (
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"github.com/mycontroller-org/server/v2/pkg/version"
	"go.uber.org/zap"
)

// Init logger
func Init() {
	logger := loggerUtils.GetLogger(cfg.CFG.Logger.Mode, cfg.CFG.Logger.Level.Core, cfg.CFG.Logger.Encoding, false, 0)
	zap.ReplaceGlobals(logger)
	zap.L().Info("welcome to the MyController world :)")
	zap.L().Info("server detail", zap.Any("version", version.Get()), zap.Any("loggerConfig", cfg.CFG.Logger))
}
