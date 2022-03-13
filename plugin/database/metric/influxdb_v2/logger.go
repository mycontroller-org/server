package influx

import (
	"fmt"
	"strings"

	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"go.uber.org/zap"
)

const callerSkipLevel = int(2)

type myLogger struct {
	logger *zap.Logger
}

func getLogger(mode, level, encoding string, enableStacktrace bool) *myLogger {
	return &myLogger{logger: loggerUtils.GetLogger(mode, level, encoding, false, callerSkipLevel, enableStacktrace)}
}

func (ml *myLogger) Sync() {
	err := ml.logger.Sync()
	if err != nil {
		zap.L().Error("error on sync", zap.Error(err))
	}
}

// SetLogLevel sets allowed logging level.
func (ml *myLogger) SetLogLevel(logLevel uint) {
	// nothing to do here, it is set on creation
}

func (ml *myLogger) LogLevel() uint {
	return 0
}

// SetPrefix sets logging prefix.
func (ml *myLogger) SetPrefix(prefix string) {
	// We are not going to use this prefix
}

// Writes formatted debug message if debug logLevel is enabled.
func (ml *myLogger) Debugf(t string, args ...interface{}) {
	ml.logger.Sugar().Debugf(fmtMsg(t), args...)
}

// Writes debug message if debug is enabled.
func (ml *myLogger) Debug(msg string) {
	ml.logger.Sugar().Debug(fmtMsg(msg))
}

// Writes formatted info message if info logLevel is enabled.
func (ml *myLogger) Infof(t string, args ...interface{}) {
	ml.logger.Sugar().Infof(fmtMsg(t), args...)
}

// Writes info message if info logLevel is enabled
func (ml *myLogger) Info(msg string) {
	ml.logger.Sugar().Info(fmtMsg(msg))
}

// Writes formatted warning message if warning logLevel is enabled.
func (ml *myLogger) Warnf(t string, args ...interface{}) {
	ml.logger.Sugar().Warnf(fmtMsg(t), args...)
}

// Writes warning message if warning logLevel is enabled.
func (ml *myLogger) Warn(msg string) {
	ml.logger.Sugar().Warn(fmtMsg(msg))
}

// Writes formatted error message
func (ml *myLogger) Errorf(t string, args ...interface{}) {
	ml.logger.Sugar().Errorf(fmtMsg(t), args...)
}

// Writes error message
func (ml *myLogger) Error(msg string) {
	ml.logger.Sugar().Error(fmtMsg(msg))
}

func fmtMsg(msg string) string {
	m := fmt.Sprintf("[MTS:INFLUXDB2CLIENT] %s", msg)
	return strings.TrimSuffix(m, "\n")
}
