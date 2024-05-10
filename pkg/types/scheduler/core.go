package schedule

import (
	"context"
	"errors"

	types "github.com/mycontroller-org/server/v2/pkg/types"
)

const (
	contextKey types.ContextKey = "scheduler"
)

type CoreScheduler interface {
	Name() string
	Start() error
	Close() error
	AddFunc(name, spec string, targetFunc func()) error
	RemoveFunc(name string)
	RemoveWithPrefix(prefix string)
	ListNames() []string
	IsAvailable(id string) bool
}

func FromContext(ctx context.Context) (CoreScheduler, error) {
	scheduler, ok := ctx.Value(contextKey).(CoreScheduler)
	if !ok {
		return nil, errors.New("invalid scheduler instance received in context")
	}
	if scheduler == nil {
		return nil, errors.New("scheduler instance not provided in context")
	}
	return scheduler, nil
}

func WithContext(ctx context.Context, scheduler CoreScheduler) context.Context {
	return context.WithValue(ctx, contextKey, scheduler)
}
