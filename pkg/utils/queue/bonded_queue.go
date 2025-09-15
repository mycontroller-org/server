// originally copied from
// https://github.com/jaegertracing/jaeger/blob/v1.55.0/pkg/queue/bounded_queue.go

// Copyright (c) 2019 The Jaeger Authors.
// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queue

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Consumer consumes data from a bounded queue
type Consumer interface {
	Consume(item interface{}) error
}

// BoundedQueue implements a producer-consumer exchange similar to a ring buffer queue,
// where the queue is bounded and if it fills up due to slow consumers, the new items written by
// the producer force the earliest items to be dropped. The implementation is actually based on
// channels, with a special Reaper goroutine that wakes up when the queue is full and consumers
// the items from the top of the queue until its size drops back to maxSize
type BoundedQueue struct {
	workers       int
	stopWG        sync.WaitGroup
	size          atomic.Int32
	capacity      atomic.Uint32
	stopped       atomic.Uint32
	items         *chan interface{}
	onDroppedItem func(item interface{})
	factory       func() Consumer
	stopCh        chan struct{}
	retryConfig   RetryConfig
}

// RetryConfig holds configuration for retrying failed items
type RetryConfig struct {
	isEnabled bool
	maxCount  uint32
	delay     time.Duration
}

// NewBoundedQueue constructs the new queue of specified capacity, and with an optional
// callback for dropped items (e.g. useful to emit metrics).
func NewBoundedQueue(capacity int, onDroppedItem func(item interface{})) *BoundedQueue {
	queue := make(chan interface{}, capacity)
	bq := &BoundedQueue{
		onDroppedItem: onDroppedItem,
		items:         &queue,
		stopCh:        make(chan struct{}),
	}
	bq.capacity.Store(uint32(capacity))
	return bq
}

// NewBoundedQueue constructs the new queue of specified capacity, and with an optional
// callback for dropped items (e.g. useful to emit metrics).
func NewBoundedQueueWithRetry(capacity int, onDroppedItem func(item interface{}), retryMaxCount uint32, retryDelay time.Duration) *BoundedQueue {
	queue := make(chan interface{}, capacity)
	bq := &BoundedQueue{
		onDroppedItem: onDroppedItem,
		items:         &queue,
		stopCh:        make(chan struct{}),
		retryConfig: RetryConfig{
			isEnabled: true,
			maxCount:  retryMaxCount,
			delay:     retryDelay,
		},
	}
	bq.capacity.Store(uint32(capacity))
	return bq
}

// StartConsumersWithFactory creates a given number of consumers consuming items
// from the queue in separate goroutines.
func (q *BoundedQueue) StartConsumersWithFactory(num int, factory func() Consumer) {
	q.workers = num
	q.factory = factory
	var startWG sync.WaitGroup
	for i := 0; i < q.workers; i++ {
		q.stopWG.Add(1)
		startWG.Add(1)
		go func() {
			startWG.Done()
			defer q.stopWG.Done()
			consumer := q.factory()
			queue := *q.items

			if q.retryConfig.isEnabled {
				// Create a local queue for retrying items at the front
				var retryItem interface{}
				hasRetry := false
				retryDelay := time.Millisecond * 100
				retryAttemptNumber := uint32(0)

				for {
					// If we have a retry item, process it first
					if hasRetry {
						if q.retryConfig.maxCount > 0 && retryAttemptNumber >= q.retryConfig.maxCount {
							// Max retries exceeded, drop the item
							q.size.Add(-1)
							if q.onDroppedItem != nil {
								q.onDroppedItem(retryItem)
							}
							hasRetry = false
							retryItem = nil
							retryDelay = time.Millisecond * 100
							retryAttemptNumber = 0
							continue
						}

						select {
						case <-time.After(retryDelay):
							retryAttemptNumber++
							err := consumer.Consume(retryItem)
							if err != nil {
								// Still failing, increase delay (exponential backoff)
								retryDelay = retryDelay * 2
								if retryDelay > q.retryConfig.delay {
									retryDelay = q.retryConfig.delay
								}
							} else {
								// Success! Clear retry state and decrement size
								hasRetry = false
								retryItem = nil
								retryDelay = time.Millisecond * 100
								retryAttemptNumber = 0
								q.size.Add(-1)
							}
						case <-q.stopCh:
							// Queue is closing
							if hasRetry {
								q.size.Add(-1)
							}
							return
						}
					} else {
						// No retry item, get next from queue
						select {
						case item, ok := <-queue:
							if ok {
								err := consumer.Consume(item)
								if err != nil {
									// Failed, set as retry item
									retryItem = item
									hasRetry = true
								} else {
									// Success, decrement size
									q.size.Add(-1)
								}
							} else {
								// channel closed, finish worker
								return
							}
						case <-q.stopCh:
							// the whole queue is closing, finish worker
							return
						}
					}
				}
			} else {
				for {
					select {
					case item, ok := <-queue:
						if ok {
							_ = consumer.Consume(item)
							q.size.Add(-1)
						} else {
							// channel closed, finish worker
							return
						}
					case <-q.stopCh:
						// the whole queue is closing, finish worker
						return
					}
				}
			}
		}()
	}
	startWG.Wait()
}

// ConsumerFunc is an adapter to allow the use of
// a consume function callback as a Consumer.
type ConsumerFunc func(item interface{}) error

// Consume calls c(item)
func (c ConsumerFunc) Consume(item interface{}) error {
	return c(item)
}

// StartConsumers starts a given number of goroutines consuming items from the queue
// and passing them into the consumer callback.
func (q *BoundedQueue) StartConsumers(num int, callback func(item interface{}) error) {
	q.StartConsumersWithFactory(num, func() Consumer {
		return ConsumerFunc(callback)
	})
}

// StopConsumers stops all the consumers
// should be started again with StartConsumers func
func (q *BoundedQueue) StopConsumers() {
	q.stopWG.Done()
}

// Produce is used by the producer to submit new item to the queue. Returns false in case of queue overflow.
func (q *BoundedQueue) Produce(item interface{}) bool {
	if q.stopped.Load() != 0 {
		if q.onDroppedItem != nil {
			q.onDroppedItem(item)
		}
		return false
	}

	// we might have two concurrent backing queues at the moment
	// their combined size is stored in q.size, and their combined capacity
	// should match the capacity of the new queue
	if q.Size() >= q.Capacity() {
		// note that all items will be dropped if the capacity is 0
		if q.onDroppedItem != nil {
			q.onDroppedItem(item)
		}
		return false
	}

	q.size.Add(1)
	select {
	case *q.items <- item:
		return true
	default:
		// should not happen, as overflows should have been captured earlier
		q.size.Add(-1)
		if q.onDroppedItem != nil {
			q.onDroppedItem(item)
		}
		return false
	}
}

// Stop stops all consumers, as well as the length reporter if started,
// and releases the items channel. It blocks until all consumers have stopped.
func (q *BoundedQueue) Stop() {
	// Use atomic CAS to ensure Stop is only executed once
	if q.stopped.CompareAndSwap(0, 1) {
		close(q.stopCh)
		q.stopWG.Wait()
		close(*q.items)
	}
}

// Size returns the current size of the queue
func (q *BoundedQueue) Size() int {
	return int(q.size.Load())
}

// Capacity returns capacity of the queue
func (q *BoundedQueue) Capacity() int {
	return int(q.capacity.Load())
}

// Resize changes the capacity of the queue, returning whether the action was successful
func (q *BoundedQueue) Resize(capacity int) bool {
	if capacity == q.Capacity() {
		// noop
		return false
	}

	previous := *q.items
	queue := make(chan interface{}, capacity)

	// swap queues
	// #nosec
	swapped := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.items)), unsafe.Pointer(q.items), unsafe.Pointer(&queue))
	if swapped {
		// start a new set of consumers, based on the information given previously
		q.StartConsumersWithFactory(q.workers, q.factory)

		// gracefully drain the existing queue
		close(previous)

		// update the capacity
		q.capacity.Store(uint32(capacity))
	}

	return swapped
}
