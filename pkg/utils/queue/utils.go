package queue

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// Queue to hold items
type Queue struct {
	Name    string
	Queue   *BoundedQueue
	Limit   int
	Workers int
	closed  concurrency.SafeBool
}

// New returns brandnew queue
func New(logger *zap.Logger, name string, limit int, consumer func(item interface{}) error, workers int) *Queue {
	// Enable retry by default with unlimited retries (0 means unlimited) and 5 second max delay
	return NewWithRetry(logger, name, limit, consumer, workers, true, 0, 5*time.Second)
}

// New returns brandnew queue with retry options
func NewWithRetry(logger *zap.Logger, name string, limit int, consumer func(item interface{}) error, workers int, isRetryEnabled bool, maxRetryCount uint32, retryDelay time.Duration) *Queue {
	droppedItemHandler := func(item interface{}) {
		logger.Error("queue full. dropping item", zap.String("queueName", name), zap.Any("item", item))
	}

	var queue *BoundedQueue
	if isRetryEnabled {
		queue = NewBoundedQueueWithRetry(limit, droppedItemHandler, maxRetryCount, retryDelay)
	} else {
		queue = NewBoundedQueue(limit, droppedItemHandler)
	}

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

// used to hold queue and subscription details
type QueueSpec struct {
	Topic          string
	SubscriptionId int64
	Queue          *Queue
}

func (qs *QueueSpec) Close() {
	qs.Queue.Close()
}

func (qs *QueueSpec) Produce(item interface{}) bool {
	return qs.Queue.Produce(item)
}

func (qs *QueueSpec) Size() int {
	return qs.Queue.Size()
}
