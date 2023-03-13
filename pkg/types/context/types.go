package types

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

const (
	MC_API_CONTEXT ContextKey = "mc_api_context"
	MC_CONFIG      ContextKey = "mc_config"
	LOGGER         ContextKey = "mc_logger"
	SCHEDULER      ContextKey = "mc_scheduler"
	STORAGE_DB     ContextKey = "mc_storage_db"
	METRIC_DB      ContextKey = "mc_metric_db"
	BUS            ContextKey = "mc_bus"
	ENTITY_API     ContextKey = "mc_entity_api"
	ENCRYPTION_API ContextKey = "mc_encryption_api"
)

type ContextKey string

// struct used in api request
type McApiContext struct {
	Tenant string `json:"tenant" yaml:"tenant"`
	UserID string `json:"userId" yaml:"userId"`
}

func LoggerFromContext(ctx context.Context) (*zap.Logger, error) {
	logger, ok := ctx.Value(LOGGER).(*zap.Logger)
	if !ok {
		return nil, errors.New("invalid logger instance received in context")
	}
	if logger == nil {
		return nil, errors.New("logger instance not provided in context")
	}
	return logger, nil
}

func LoggerWithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, LOGGER, logger)
}
