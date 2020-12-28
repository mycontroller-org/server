package embedded

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const loggerPrefix = "[BUS:EMBEDDED]"

// PrintDebug log
func PrintDebug(message string, fields ...zapcore.Field) {
	zap.L().Debug(fmt.Sprintf("%s %s", loggerPrefix, message), fields...)
}

// PrintInfo log
func PrintInfo(message string, fields ...zapcore.Field) {
	zap.L().Info(fmt.Sprintf("%s %s", loggerPrefix, message), fields...)
}

// PrintWarn log
func PrintWarn(message string, fields ...zapcore.Field) {
	zap.L().Warn(fmt.Sprintf("%s %s", loggerPrefix, message), fields...)
}

// PrintError log
func PrintError(message string, fields ...zapcore.Field) {
	zap.L().Error(fmt.Sprintf("%s %s", loggerPrefix, message), fields...)
}
