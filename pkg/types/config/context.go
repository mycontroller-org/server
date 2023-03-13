package config

import (
	"context"
	"errors"

	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
)

func FrosmContext(ctx context.Context) (*Config, error) {
	cfg, ok := ctx.Value(contextTY.MC_CONFIG).(*Config)
	if !ok {
		return nil, errors.New("invalid or config instance not available in context")
	}
	if cfg == nil {
		return nil, errors.New("config instance not provided in context")
	}
	return cfg, nil
}

func WithContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextTY.MC_CONFIG, cfg)
}
