package utils

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ModeRecordAll = "record_all"
	ModeSampled   = "sampled"
)

// GetLogger returns a logger
func GetLogger(mode, level, encoding string, showFullCaller bool, callerSkip int) *zap.Logger {
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

	if showFullCaller {
		zapCfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}
	// update user change
	// update log level
	switch strings.ToLower(level) {
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

	// print the logs on Stdout, default "Stderr"
	zapCfg.OutputPaths = []string{"Stdout"}

	logger, err := zapCfg.Build(zap.AddCaller(), zap.AddCallerSkip(callerSkip))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return logger
}
