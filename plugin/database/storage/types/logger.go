package storage

import (
	"github.com/mycontroller-org/server/v2/pkg/types"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"go.uber.org/zap"
)

// returns a logger with environment values
func GetStorageLogger() *zap.Logger {
	return loggerUtils.GetLogger(
		types.GetEnvString(types.ENV_LOG_MODE),
		types.GetEnvString(types.ENV_LOG_LEVEL_STORAGE),
		types.GetEnvString(types.ENV_LOG_ENCODING),
		false,
		0,
		types.GetEnvBool(types.ENV_LOG_ENABLE_STACK_TRACE),
	)
}
