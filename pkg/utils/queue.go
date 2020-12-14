package utils

import (
	q "github.com/jaegertracing/jaeger/pkg/queue"
	"go.uber.org/zap"
)

// GetQueue return the queue with the requested size
func GetQueue(name string, size int) *q.BoundedQueue {
	return q.NewBoundedQueue(size, func(item interface{}) {
		zap.L().Error("Queue full. Droping item", zap.String("name", name), zap.Any("item", item))
	})
}
