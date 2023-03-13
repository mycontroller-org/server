package javascript

import (
	"github.com/dop251/goja_nodejs/console"
	"go.uber.org/zap"
)

type customLogger struct {
	logger *zap.Logger
}

func getCustomConsoleLogger(logger *zap.Logger) console.Printer {
	return &customLogger{logger: logger}
}

func (cl *customLogger) Log(msg string) {
	cl.logger.Info(msg)
}

func (cl *customLogger) Warn(msg string) {
	cl.logger.Warn(msg)
}

func (cl *customLogger) Error(msg string) {
	cl.logger.Error(msg)
}
