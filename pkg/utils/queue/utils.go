package queue

import (
	queue "github.com/jaegertracing/jaeger/pkg/queue"
	"go.uber.org/zap"
)

// Queue to hold items
type Queue struct {
	Name  string
	Queue *queue.BoundedQueue
	Size  int
}

// New returns brandnew queue
func New(name string, size int, consumer func(event interface{}), consumersWorkers int) *Queue {
	queue := queue.NewBoundedQueue(size, func(item interface{}) {
		zap.L().Error("Queue full. Droping item", zap.String("QueueName", name), zap.Any("item", item))
	})

	queue.StartConsumers(consumersWorkers, consumer)

	return &Queue{
		Name:  name,
		Size:  size,
		Queue: queue,
	}
}

// Close the queue
func (q *Queue) Close() {
	q.Queue.Stop()
}

// Produce adds an item to the queue
func (q *Queue) Produce(item interface{}) bool {
	return q.Queue.Produce(item)
}
