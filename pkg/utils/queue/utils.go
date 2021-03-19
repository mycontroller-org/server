package queue

import (
	queue "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// Queue to hold items
type Queue struct {
	Name    string
	Queue   *queue.BoundedQueue
	Limit   int
	Workers int
	closed  concurrency.SafeBool
}

// New returns brandnew queue
func New(name string, limit int, consumer func(event interface{}), workers int) *Queue {
	droppedItemHandler := func(item interface{}) {
		zap.L().Error("Queue full. Droping item", zap.String("QueueName", name), zap.Any("item", item))
	}

	queue := queue.NewBoundedQueue(limit, droppedItemHandler)
	queue.StartConsumers(workers, consumer)

	return &Queue{
		Name:    name,
		Queue:   queue,
		Limit:   limit,
		Workers: workers,
		closed:  concurrency.SafeBool{},
	}
}

// Close the queue
func (q *Queue) Close() {
	if !q.closed.IsSet() {
		q.Queue.Stop()
		q.closed.Set()
	}
}

// Produce adds an item to the queue
func (q *Queue) Produce(item interface{}) bool {
	if q.closed.IsSet() {
		return false
	}
	return q.Queue.Produce(item)
}

// Size returns current size of the queue
func (q *Queue) Size() int {
	return q.Queue.Size()
}
