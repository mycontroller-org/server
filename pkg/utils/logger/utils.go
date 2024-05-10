package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ModeRecordAll                  = "record_all"
	ModeSampled                    = "sampled"
	contextKey    types.ContextKey = "logger"
)

// GetLogger returns a logger
func GetLogger(mode, logLevel, encoding string, showFullCaller bool, callerSkip int, enableStacktrace bool) *zap.Logger {
	var zapCfg zap.Config
	if strings.ToLower(mode) == ModeSampled {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	zapCfg.EncoderConfig.TimeKey = "time"
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.NameKey = "logger"
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.MessageKey = "msg"
	zapCfg.EncoderConfig.StacktraceKey = "stacktrace"
	zapCfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.DisableStacktrace = !enableStacktrace

	if showFullCaller {
		zapCfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}
	// update user change
	// update log level
	switch strings.ToLower(logLevel) {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn", "warning":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	// update encoding type
	switch strings.ToLower(encoding) {
	case "json":
		zapCfg.Encoding = "json"
	default:
		zapCfg.Encoding = "console"
	}

	// print the logs on 'stdout', default 'stderr'
	zapCfg.OutputPaths = []string{"stdout"}

	logger, err := zapCfg.Build(zap.AddCaller(), zap.AddCallerSkip(callerSkip))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return logger
}

func FromContext(ctx context.Context) (*zap.Logger, error) {
	logger, ok := ctx.Value(contextKey).(*zap.Logger)
	if !ok {
		return nil, errors.New("invalid logger instance received in context")
	}
	if logger == nil {
		return nil, errors.New("logger instance not provided in context")
	}
	return logger, nil
}

func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}
