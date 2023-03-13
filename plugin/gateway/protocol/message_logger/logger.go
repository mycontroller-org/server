package msglogger

import (
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"go.uber.org/zap"
)

// MessageLogger interface
type MessageLogger interface {
	Start()                              // Start the message logger
	Close()                              // Stop and close the message logger
	AsyncWrite(rawMsg *msgTY.RawMessage) // write a message
}

// logger types
const (
	TypeVoidLogger = "void_logger"
	TypeFileLogger = "file_logger"
)

// New message logger
func New(logger *zap.Logger, gatewayID string, config cmap.CustomMap, formatterFunc func(rawMsg *msgTY.RawMessage) string, logRootDir string) MessageLogger {
	var messageLogger MessageLogger
	if config.GetString(types.NameType) == TypeFileLogger {
		fileMessageLogger, err := NewFileMessageLogger(logger, gatewayID, config, formatterFunc, logRootDir)
		if err != nil {
			logger.Error("Failed to load file message logger", zap.Any("config", config), zap.Error(err))
		} else {
			messageLogger = fileMessageLogger
		}
	}
	// if non loaded load void logger
	if messageLogger == nil {
		messageLogger = &NoopMessageLogger{}
		logger.Debug("Loaded void logger", zap.String("gateway", gatewayID), zap.Any("config", config))
	}
	return messageLogger
}

// GetNoopLogger can be used for pre stage
func GetNoopLogger() MessageLogger {
	return &NoopMessageLogger{}
}
