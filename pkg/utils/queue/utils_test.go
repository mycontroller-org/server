package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestNew tests the basic queue creation function
func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		queueName   string
		limit       int
		workers     int
		wantLimit   int
		wantWorkers int
		wantName    string
	}{
		{
			name:        "basic queue",
			queueName:   "test_queue",
			limit:       100,
			workers:     2,
			wantLimit:   100,
			wantWorkers: 2,
			wantName:    "test_queue",
		},
		{
			name:        "single worker",
			queueName:   "single",
			limit:       10,
			workers:     1,
			wantLimit:   10,
			wantWorkers: 1,
			wantName:    "single",
		},
		{
			name:        "high capacity",
			queueName:   "high_cap",
			limit:       10000,
			workers:     10,
			wantLimit:   10000,
			wantWorkers: 10,
			wantName:    "high_cap",
		},
		{
			name:        "zero workers",
			queueName:   "zero_workers",
			limit:       50,
			workers:     0,
			wantLimit:   50,
			wantWorkers: 0,
			wantName:    "zero_workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			var processed []interface{}
			var mu sync.Mutex

			consumer := func(item interface{}) error {
				mu.Lock()
				processed = append(processed, item)
				mu.Unlock()
				return nil
			}

			q := New(logger, tt.queueName, tt.limit, consumer, tt.workers)
			defer q.Close()

			if q.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", q.Name, tt.wantName)
			}
			if q.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", q.Limit, tt.wantLimit)
			}
			if q.Workers != tt.wantWorkers {
				t.Errorf("Workers = %d, want %d", q.Workers, tt.wantWorkers)
			}
			if q.Queue == nil {
				t.Error("Queue should not be nil")
			}

			// Test basic functionality if workers > 0
			if tt.workers > 0 {
				time.Sleep(10 * time.Millisecond)
				q.Produce("test_item")
				time.Sleep(100 * time.Millisecond)

				mu.Lock()
				if len(processed) != 1 {
					t.Errorf("Expected 1 processed item, got %d", len(processed))
				}
				mu.Unlock()
			}
		})
	}
}

// TestNewWithRetry tests queue creation with retry options
func TestNewWithRetry(t *testing.T) {
	tests := []struct {
		name           string
		queueName      string
		limit          int
		workers        int
		retryEnabled   bool
		maxRetryCount  uint32
		retryDelay     time.Duration
		expectRetries  bool
		failureCount   int
		expectedFinalAttempts int
	}{
		{
			name:          "retry enabled with limited retries",
			queueName:     "retry_limited",
			limit:         10,
			workers:       1,
			retryEnabled:  true,
			maxRetryCount: 3,
			retryDelay:    10 * time.Millisecond,
			expectRetries: true,
			failureCount:  2, // fail first 2, succeed on 3rd
			expectedFinalAttempts: 3,
		},
		{
			name:          "retry enabled unlimited",
			queueName:     "retry_unlimited",
			limit:         10,
			workers:       1,
			retryEnabled:  true,
			maxRetryCount: 0, // unlimited
			retryDelay:    10 * time.Millisecond,
			expectRetries: true,
			failureCount:  4, // fail first 4, succeed on 5th
			expectedFinalAttempts: 5,
		},
		{
			name:          "retry disabled",
			queueName:     "no_retry",
			limit:         10,
			workers:       1,
			retryEnabled:  false,
			maxRetryCount: 0,
			retryDelay:    0,
			expectRetries: false,
			failureCount:  10, // always fail
			expectedFinalAttempts: 1,
		},
		{
			name:          "retry exhausted",
			queueName:     "retry_exhausted",
			limit:         10,
			workers:       1,
			retryEnabled:  true,
			maxRetryCount: 2,
			retryDelay:    10 * time.Millisecond,
			expectRetries: true,
			failureCount:  10, // always fail
			expectedFinalAttempts: 3, // 1 initial + 2 retries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			var attempts int32

			consumer := func(item interface{}) error {
				attempt := atomic.AddInt32(&attempts, 1)
				if int(attempt) <= tt.failureCount {
					return fmt.Errorf("simulated failure %d", attempt)
				}
				return nil
			}

			q := NewWithRetry(logger, tt.queueName, tt.limit, consumer, tt.workers,
				tt.retryEnabled, tt.maxRetryCount, tt.retryDelay)
			defer q.Close()

			// Verify queue properties
			if q.Name != tt.queueName {
				t.Errorf("Name = %s, want %s", q.Name, tt.queueName)
			}
			if q.Limit != tt.limit {
				t.Errorf("Limit = %d, want %d", q.Limit, tt.limit)
			}
			if q.Workers != tt.workers {
				t.Errorf("Workers = %d, want %d", q.Workers, tt.workers)
			}

			if tt.workers > 0 {
				time.Sleep(10 * time.Millisecond)
				q.Produce("test_retry")

				// Wait for retries to complete
				maxWait := time.Duration(tt.expectedFinalAttempts) * tt.retryDelay * 10
				if maxWait < 500*time.Millisecond {
					maxWait = 500 * time.Millisecond
				}
				time.Sleep(maxWait)

				finalAttempts := atomic.LoadInt32(&attempts)
				if int(finalAttempts) != tt.expectedFinalAttempts {
					t.Errorf("attempts = %d, want %d", finalAttempts, tt.expectedFinalAttempts)
				}
			}
		})
	}
}

// TestQueueOperations tests Queue wrapper methods
func TestQueueOperations(t *testing.T) {
	tests := []struct {
		name           string
		capacity       int
		items          []interface{}
		expectedProduced int
		expectedSize   int
	}{
		{
			name:            "within capacity",
			capacity:        10,
			items:           []interface{}{1, 2, 3, 4, 5},
			expectedProduced: 5,
			expectedSize:    5,
		},
		{
			name:            "at capacity",
			capacity:        3,
			items:           []interface{}{"a", "b", "c"},
			expectedProduced: 3,
			expectedSize:    3,
		},
		{
			name:            "over capacity",
			capacity:        2,
			items:           []interface{}{1, 2, 3, 4, 5},
			expectedProduced: 2,
			expectedSize:    2,
		},
		{
			name:            "zero capacity",
			capacity:        0,
			items:           []interface{}{1, 2},
			expectedProduced: 0,
			expectedSize:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			blockChan := make(chan struct{})

			consumer := func(item interface{}) error {
				<-blockChan // Block processing to test queue filling
				return nil
			}

			q := New(logger, "test", tt.capacity, consumer, 1)
			defer func() {
				close(blockChan) // Unblock before closing
				q.Close()
			}()

			time.Sleep(10 * time.Millisecond) // Let consumers start

			produced := 0
			for _, item := range tt.items {
				if q.Produce(item) {
					produced++
				}
			}

			if produced != tt.expectedProduced {
				t.Errorf("produced = %d, want %d", produced, tt.expectedProduced)
			}

			size := q.Size()
			if size != tt.expectedSize {
				t.Errorf("Size() = %d, want %d", size, tt.expectedSize)
			}
		})
	}
}

// TestQueueClose tests queue closing behavior
func TestQueueClose(t *testing.T) {
	tests := []struct {
		name        string
		capacity    int
		workers     int
		closeFirst  bool
	}{
		{"normal close", 10, 2, false},
		{"close before produce", 10, 2, true},
		{"single worker close", 5, 1, false},
		{"multiple workers close", 20, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			consumer := func(item interface{}) error { return nil }

			q := New(logger, "test", tt.capacity, consumer, tt.workers)

			if tt.closeFirst {
				q.Close()
				// Should not be able to produce after close
				if q.Produce("after_close") {
					t.Error("Should not be able to produce after close")
				}
			} else {
				time.Sleep(10 * time.Millisecond)
				// Should be able to produce before close
				if !q.Produce("before_close") {
					t.Error("Should be able to produce before close")
				}
				q.Close()
				// Should not be able to produce after close
				if q.Produce("after_close") {
					t.Error("Should not be able to produce after close")
				}
			}

			// Multiple closes should be safe
			q.Close()
			q.Close()
		})
	}
}

// TestQueueSpec tests the QueueSpec wrapper
func TestQueueSpec(t *testing.T) {
	tests := []struct {
		name           string
		topic          string
		subscriptionId int64
		capacity       int
		items          []interface{}
	}{
		{
			name:           "basic queue spec",
			topic:          "test_topic",
			subscriptionId: 12345,
			capacity:       10,
			items:          []interface{}{"msg1", "msg2", "msg3"},
		},
		{
			name:           "different topic",
			topic:          "another_topic",
			subscriptionId: 67890,
			capacity:       5,
			items:          []interface{}{1, 2},
		},
		{
			name:           "empty topic",
			topic:          "",
			subscriptionId: 0,
			capacity:       1,
			items:          []interface{}{"single"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			var processed []interface{}
			var mu sync.Mutex

			consumer := func(item interface{}) error {
				mu.Lock()
				processed = append(processed, item)
				mu.Unlock()
				return nil
			}

			q := New(logger, "test", tt.capacity, consumer, 1)
			qs := &QueueSpec{
				Topic:          tt.topic,
				SubscriptionId: tt.subscriptionId,
				Queue:          q,
			}
			defer qs.Close()

			// Verify properties
			if qs.Topic != tt.topic {
				t.Errorf("Topic = %s, want %s", qs.Topic, tt.topic)
			}
			if qs.SubscriptionId != tt.subscriptionId {
				t.Errorf("SubscriptionId = %d, want %d", qs.SubscriptionId, tt.subscriptionId)
			}

			time.Sleep(10 * time.Millisecond)

			// Test operations through QueueSpec
			for _, item := range tt.items {
				if !qs.Produce(item) {
					t.Errorf("Failed to produce item %v", item)
				}
			}

			time.Sleep(100 * time.Millisecond)

			// Verify processing
			mu.Lock()
			if len(processed) != len(tt.items) {
				t.Errorf("processed count = %d, want %d", len(processed), len(tt.items))
			}
			mu.Unlock()

			// Test size method
			size := qs.Size()
			if size < 0 {
				t.Errorf("Size should not be negative, got %d", size)
			}
		})
	}
}

// TestQueueConcurrency tests concurrent operations on Queue wrapper
func TestQueueConcurrency(t *testing.T) {
	tests := []struct {
		name         string
		capacity     int
		workers      int
		producers    int
		itemsEach    int
		consumerDelay time.Duration
	}{
		{
			name:         "light concurrent load",
			capacity:     100,
			workers:      2,
			producers:    3,
			itemsEach:    20,
			consumerDelay: time.Millisecond,
		},
		{
			name:         "heavy concurrent load",
			capacity:     500,
			workers:      5,
			producers:    10,
			itemsEach:    50,
			consumerDelay: 0,
		},
		{
			name:         "single worker many producers",
			capacity:     200,
			workers:      1,
			producers:    8,
			itemsEach:    25,
			consumerDelay: time.Microsecond * 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			var consumed int64
			var successful int64

			consumer := func(item interface{}) error {
				if tt.consumerDelay > 0 {
					time.Sleep(tt.consumerDelay)
				}
				atomic.AddInt64(&consumed, 1)
				return nil
			}

			q := New(logger, "concurrent_test", tt.capacity, consumer, tt.workers)
			defer q.Close()

			time.Sleep(10 * time.Millisecond) // Let consumers start

			var wg sync.WaitGroup

			// Start producers
			for p := 0; p < tt.producers; p++ {
				wg.Add(1)
				go func(producerId int) {
					defer wg.Done()
					for i := 0; i < tt.itemsEach; i++ {
						item := fmt.Sprintf("producer_%d_item_%d", producerId, i)
						if q.Produce(item) {
							atomic.AddInt64(&successful, 1)
						}
					}
				}(p)
			}

			// Monitor size concurrently
			ctx, cancel := context.WithCancel(context.Background())
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						size := q.Size()
						if size < 0 {
							t.Errorf("Size should not be negative: %d", size)
							return
						}
						if size > tt.capacity {
							t.Errorf("Size %d should not exceed capacity %d", size, tt.capacity)
							return
						}
						time.Sleep(time.Millisecond)
					}
				}
			}()

			wg.Wait()
			cancel()

			// Wait for processing to complete
			time.Sleep(time.Duration(tt.workers) * 100 * time.Millisecond)

			finalSuccessful := atomic.LoadInt64(&successful)
			finalConsumed := atomic.LoadInt64(&consumed)

			if finalConsumed != finalSuccessful {
				t.Errorf("consumed = %d, successful = %d, should be equal", finalConsumed, finalSuccessful)
			}

			// Final size should be reasonable
			finalSize := q.Size()
			if finalSize < 0 {
				t.Errorf("Final size should not be negative: %d", finalSize)
			}
		})
	}
}

// TestQueueErrorScenarios tests various error conditions
func TestQueueErrorScenarios(t *testing.T) {
	tests := []struct {
		name              string
		capacity          int
		workers           int
		consumerFunc      func(item interface{}) error
		expectProcessing  bool
		testPanic         bool
	}{
		{
			name:     "nil consumer",
			capacity: 10,
			workers:  1,
			consumerFunc: nil,
			expectProcessing: false,
			testPanic: false,
		},
		{
			name:     "consumer with error",
			capacity: 10,
			workers:  1,
			consumerFunc: func(item interface{}) error {
				return errors.New("consumer error")
			},
			expectProcessing: true, // Items processed but with errors
			testPanic: false,
		},
		{
			name:     "consumer with panic",
			capacity: 10,
			workers:  1,
			consumerFunc: func(item interface{}) error {
				if item == "panic_item" {
					panic("test panic")
				}
				return nil
			},
			expectProcessing: true,
			testPanic: true,
		},
		{
			name:     "zero workers",
			capacity: 10,
			workers:  0,
			consumerFunc: func(item interface{}) error {
				return nil
			},
			expectProcessing: false,
			testPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			q := New(logger, "error_test", tt.capacity, tt.consumerFunc, tt.workers)
			defer q.Close()

			time.Sleep(10 * time.Millisecond)

			if tt.testPanic {
				// This should not crash the test
				q.Produce("panic_item")
				time.Sleep(100 * time.Millisecond)

				// Queue should still be functional
				q.Produce("normal_item")
				time.Sleep(100 * time.Millisecond)
			} else {
				q.Produce("test_item")
				time.Sleep(100 * time.Millisecond)
			}

			size := q.Size()
			if !tt.expectProcessing && tt.workers > 0 {
				// Items should remain in queue if not processing
				if size <= 0 {
					t.Errorf("Expected items to remain in queue, size = %d", size)
				}
			}
		})
	}
}

// TestQueueMemoryManagement tests proper cleanup and resource management
func TestQueueMemoryManagement(t *testing.T) {
	t.Run("multiple queue creation and cleanup", func(t *testing.T) {
		logger := zap.NewNop()
		consumer := func(item interface{}) error { return nil }

		queues := make([]*Queue, 10)

		// Create multiple queues
		for i := 0; i < 10; i++ {
			queues[i] = New(logger, fmt.Sprintf("queue_%d", i), 10, consumer, 1)

			// Add some items
			queues[i].Produce(fmt.Sprintf("item_%d", i))
		}

		time.Sleep(50 * time.Millisecond)

		// Close all queues
		for _, q := range queues {
			q.Close()

			// Verify operations fail after close
			if q.Produce("after_close") {
				t.Error("Produce should fail after close")
			}
		}
	})

	t.Run("queue spec cleanup", func(t *testing.T) {
		logger := zap.NewNop()
		consumer := func(item interface{}) error { return nil }

		queueSpecs := make([]*QueueSpec, 5)

		for i := 0; i < 5; i++ {
			q := New(logger, fmt.Sprintf("spec_queue_%d", i), 5, consumer, 1)
			queueSpecs[i] = &QueueSpec{
				Topic:          fmt.Sprintf("topic_%d", i),
				SubscriptionId: int64(i + 100),
				Queue:          q,
			}

			queueSpecs[i].Produce(fmt.Sprintf("spec_item_%d", i))
		}

		time.Sleep(50 * time.Millisecond)

		for _, qs := range queueSpecs {
			qs.Close()

			if qs.Produce("after_spec_close") {
				t.Error("QueueSpec Produce should fail after close")
			}
		}
	})
}

// BenchmarkQueueNew benchmarks queue creation
func BenchmarkQueueNew(b *testing.B) {
	logger := zap.NewNop()
	consumer := func(item interface{}) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := New(logger, "bench", 100, consumer, 2)
		q.Close()
	}
}

// BenchmarkQueueProduce benchmarks queue production through wrapper
func BenchmarkQueueProduce(b *testing.B) {
	logger := zap.NewNop()
	consumer := func(item interface{}) error { return nil }
	q := New(logger, "bench", 10000, consumer, 4)
	defer q.Close()

	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Produce(i)
	}
}

// BenchmarkQueueSpecProduce benchmarks production through QueueSpec
func BenchmarkQueueSpecProduce(b *testing.B) {
	logger := zap.NewNop()
	consumer := func(item interface{}) error { return nil }
	q := New(logger, "bench", 10000, consumer, 4)
	qs := &QueueSpec{
		Topic:          "benchmark",
		SubscriptionId: 1,
		Queue:          q,
	}
	defer qs.Close()

	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.Produce(i)
	}
}

// BenchmarkConcurrentQueueAccess benchmarks concurrent access
func BenchmarkConcurrentQueueAccess(b *testing.B) {
	logger := zap.NewNop()
	consumer := func(item interface{}) error { return nil }
	q := New(logger, "bench", 1000, consumer, 4)
	defer q.Close()

	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			q.Produce(i)
			i++
		}
	})
}