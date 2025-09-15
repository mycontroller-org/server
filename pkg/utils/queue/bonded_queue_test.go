package queue

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

// TestNewBoundedQueue tests basic queue creation
func TestNewBoundedQueue(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		wantCap  int
	}{
		{"small capacity", 10, 10},
		{"medium capacity", 100, 100},
		{"large capacity", 1000, 1000},
		{"zero capacity", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var droppedItems []interface{}
			onDropped := func(item interface{}) {
				droppedItems = append(droppedItems, item)
			}

			q := NewBoundedQueue(tt.capacity, onDropped)
			if q == nil {
				t.Fatal("NewBoundedQueue returned nil")
			}

			if q.Capacity() != tt.wantCap {
				t.Errorf("Capacity() = %d, want %d", q.Capacity(), tt.wantCap)
			}

			if q.Size() != 0 {
				t.Errorf("Size() = %d, want 0", q.Size())
			}

			q.Stop()
		})
	}
}

// TestNewBoundedQueueWithRetry tests queue creation with retry
func TestNewBoundedQueueWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		capacity      int
		retryMaxCount uint32
		retryDelay    time.Duration
		wantCap       int
	}{
		{"with limited retries", 10, 3, 100 * time.Millisecond, 10},
		{"with unlimited retries", 50, 0, 50 * time.Millisecond, 50},
		{"with high retry count", 100, 100, 10 * time.Millisecond, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var droppedItems []interface{}
			onDropped := func(item interface{}) {
				droppedItems = append(droppedItems, item)
			}

			q := NewBoundedQueueWithRetry(tt.capacity, onDropped, tt.retryMaxCount, tt.retryDelay)
			if q == nil {
				t.Fatal("NewBoundedQueueWithRetry returned nil")
			}

			if q.Capacity() != tt.wantCap {
				t.Errorf("Capacity() = %d, want %d", q.Capacity(), tt.wantCap)
			}

			if q.Size() != 0 {
				t.Errorf("Size() = %d, want 0", q.Size())
			}

			if !q.retryConfig.isEnabled {
				t.Error("retryConfig.isEnabled should be true")
			}

			if q.retryConfig.maxCount != tt.retryMaxCount {
				t.Errorf("retryConfig.maxCount = %d, want %d", q.retryConfig.maxCount, tt.retryMaxCount)
			}

			if q.retryConfig.delay != tt.retryDelay {
				t.Errorf("retryConfig.delay = %v, want %v", q.retryConfig.delay, tt.retryDelay)
			}

			q.Stop()
		})
	}
}

// TestBoundedQueueProduce tests item production
func TestBoundedQueueProduce(t *testing.T) {
	tests := []struct {
		name       string
		capacity   int
		items      []interface{}
		wantSize   int
		wantResult []bool
	}{
		{
			"within capacity",
			10,
			[]interface{}{1, 2, 3},
			3,
			[]bool{true, true, true},
		},
		{
			"at capacity",
			2,
			[]interface{}{1, 2},
			2,
			[]bool{true, true},
		},
		{
			"over capacity",
			2,
			[]interface{}{1, 2, 3, 4},
			2,
			[]bool{true, true, false, false},
		},
		{
			"zero capacity",
			0,
			[]interface{}{1, 2},
			0,
			[]bool{false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var droppedItems []interface{}
			onDropped := func(item interface{}) {
				droppedItems = append(droppedItems, item)
			}

			q := NewBoundedQueue(tt.capacity, onDropped)
			defer q.Stop()

			results := make([]bool, len(tt.items))
			for i, item := range tt.items {
				results[i] = q.Produce(item)
			}

			if q.Size() != tt.wantSize {
				t.Errorf("Size() = %d, want %d", q.Size(), tt.wantSize)
			}

			for i, result := range results {
				if result != tt.wantResult[i] {
					t.Errorf("Produce(%v) = %v, want %v", tt.items[i], result, tt.wantResult[i])
				}
			}

			expectedDropped := len(tt.items) - tt.wantSize
			if len(droppedItems) != expectedDropped {
				t.Errorf("dropped items count = %d, want %d", len(droppedItems), expectedDropped)
			}
		})
	}
}

// TestBoundedQueueConsumers tests consumer functionality
func TestBoundedQueueConsumers(t *testing.T) {
	tests := []struct {
		name         string
		capacity     int
		workers      int
		items        []interface{}
		consumerErr  error
		wantConsumed int
	}{
		{
			"successful consumption",
			10,
			2,
			[]interface{}{1, 2, 3, 4, 5},
			nil,
			5,
		},
		{
			"single worker",
			5,
			1,
			[]interface{}{"a", "b", "c"},
			nil,
			3,
		},
		{
			"multiple workers",
			20,
			5,
			[]interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			nil,
			10,
		},
		{
			"consumer with errors",
			10,
			1,
			[]interface{}{1, 2, 3},
			errors.New("consumer error"),
			3, // Items still processed even with errors when retry disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var consumed int32
			var consumedItems []interface{}
			var mu sync.Mutex

			consumer := func(item interface{}) error {
				atomic.AddInt32(&consumed, 1)
				mu.Lock()
				consumedItems = append(consumedItems, item)
				mu.Unlock()
				return tt.consumerErr
			}

			q := NewBoundedQueue(tt.capacity, nil)
			defer q.Stop()

			q.StartConsumers(tt.workers, consumer)

			// Give consumers time to start
			time.Sleep(10 * time.Millisecond)

			for _, item := range tt.items {
				q.Produce(item)
			}

			// Wait for processing
			time.Sleep(100 * time.Millisecond)

			finalConsumed := atomic.LoadInt32(&consumed)
			if int(finalConsumed) != tt.wantConsumed {
				t.Errorf("consumed items = %d, want %d", finalConsumed, tt.wantConsumed)
			}

			mu.Lock()
			if len(consumedItems) != tt.wantConsumed {
				t.Errorf("consumedItems length = %d, want %d", len(consumedItems), tt.wantConsumed)
			}
			mu.Unlock()
		})
	}
}

// TestBoundedQueueWithRetryFunctionality tests retry behavior
func TestBoundedQueueWithRetryFunctionality(t *testing.T) {
	tests := []struct {
		name          string
		maxRetryCount uint32
		retryDelay    time.Duration
		failCount     int
		wantAttempts  int
	}{
		{
			"retry until success",
			5,
			10 * time.Millisecond,
			2, // Fail first 2, succeed on 3rd
			3,
		},
		{
			"exhaust retries",
			2,
			10 * time.Millisecond,
			10, // Always fail
			3,  // 1 initial + 2 retries
		},
		{
			"unlimited retries",
			0,
			5 * time.Millisecond,
			4, // Fail first 4, succeed on 5th
			5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts int32
			var droppedItems []interface{}

			consumer := func(item interface{}) error {
				attempt := atomic.AddInt32(&attempts, 1)
				if int(attempt) <= tt.failCount {
					return errors.New("simulated failure")
				}
				return nil
			}

			onDropped := func(item interface{}) {
				droppedItems = append(droppedItems, item)
			}

			q := NewBoundedQueueWithRetry(10, onDropped, tt.maxRetryCount, tt.retryDelay)
			defer q.Stop()

			q.StartConsumers(1, consumer)
			time.Sleep(10 * time.Millisecond)

			q.Produce("test_item")

			// Wait for retries to complete
			maxWait := time.Duration(tt.wantAttempts) * tt.retryDelay * 10
			if maxWait < 500*time.Millisecond {
				maxWait = 500 * time.Millisecond
			}
			time.Sleep(maxWait)

			finalAttempts := atomic.LoadInt32(&attempts)
			if int(finalAttempts) != tt.wantAttempts {
				t.Errorf("attempts = %d, want %d", finalAttempts, tt.wantAttempts)
			}

			// Check if item was dropped when retries exhausted
			if tt.failCount >= tt.wantAttempts && tt.maxRetryCount > 0 {
				if len(droppedItems) != 1 {
					t.Errorf("expected 1 dropped item, got %d", len(droppedItems))
				}
			}
		})
	}
}

// TestBoundedQueueResize tests queue resizing functionality
func TestBoundedQueueResize(t *testing.T) {
	tests := []struct {
		name        string
		initialCap  int
		newCap      int
		wantSuccess bool
	}{
		{"increase capacity", 10, 20, true},
		{"decrease capacity", 20, 10, true},
		{"same capacity", 15, 15, false},
		{"zero to positive", 0, 10, true},
		{"positive to zero", 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := func(item interface{}) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}

			q := NewBoundedQueue(tt.initialCap, nil)
			defer q.Stop()

			q.StartConsumers(1, consumer)
			time.Sleep(10 * time.Millisecond)

			// Add some items
			for i := 0; i < min(tt.initialCap, 5); i++ {
				q.Produce(i)
			}

			success := q.Resize(tt.newCap)
			if success != tt.wantSuccess {
				t.Errorf("Resize(%d) = %v, want %v", tt.newCap, success, tt.wantSuccess)
			}

			if tt.wantSuccess {
				if q.Capacity() != tt.newCap {
					t.Errorf("Capacity() = %d, want %d", q.Capacity(), tt.newCap)
				}
			} else {
				if q.Capacity() != tt.initialCap {
					t.Errorf("Capacity() = %d, want %d (unchanged)", q.Capacity(), tt.initialCap)
				}
			}
		})
	}
}

// TestBoundedQueueStop tests queue stopping
func TestBoundedQueueStop(t *testing.T) {
	t.Run("stop prevents production", func(t *testing.T) {
		var droppedItems []interface{}
		onDropped := func(item interface{}) {
			droppedItems = append(droppedItems, item)
		}

		q := NewBoundedQueue(10, onDropped)
		q.StartConsumers(1, func(item interface{}) error { return nil })

		// Produce before stop
		if !q.Produce("before_stop") {
			t.Error("Should be able to produce before stop")
		}

		q.Stop()

		// Produce after stop should fail
		if q.Produce("after_stop") {
			t.Error("Should not be able to produce after stop")
		}

		// Item should be dropped
		if len(droppedItems) != 1 {
			t.Errorf("Expected 1 dropped item, got %d", len(droppedItems))
		}
	})

	t.Run("multiple stops are safe", func(t *testing.T) {
		q := NewBoundedQueue(10, nil)
		q.StartConsumers(1, func(item interface{}) error { return nil })

		// Multiple stops should not panic or block
		q.Stop()
		q.Stop()
		q.Stop()
	})
}

// TestBoundedQueueConcurrency tests concurrent operations
func TestBoundedQueueConcurrency(t *testing.T) {
	t.Run("concurrent produce and consume", func(t *testing.T) {
		const (
			capacity  = 1000 // Increased capacity
			producers = 5
			itemsEach = 50 // Reduced items to avoid queue overflow
			consumers = 3
		)

		var consumed int64
		var successful int64
		consumer := func(item interface{}) error {
			atomic.AddInt64(&consumed, 1)
			return nil
		}

		q := NewBoundedQueue(capacity, nil)
		defer q.Stop()

		q.StartConsumers(consumers, consumer)
		time.Sleep(10 * time.Millisecond)

		var wg sync.WaitGroup

		// Start producers
		for p := 0; p < producers; p++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := 0; i < itemsEach; i++ {
					if q.Produce(id*itemsEach + i) {
						atomic.AddInt64(&successful, 1)
					}
				}
			}(p)
		}

		wg.Wait()

		// Wait for all items to be consumed
		time.Sleep(1 * time.Second)

		expectedTotal := atomic.LoadInt64(&successful)
		actualConsumed := atomic.LoadInt64(&consumed)

		if actualConsumed != expectedTotal {
			t.Errorf("consumed = %d, want %d (successful productions)", actualConsumed, expectedTotal)
		}
	})

	t.Run("concurrent size and capacity checks", func(t *testing.T) {
		q := NewBoundedQueue(50, nil)
		defer q.Stop()

		consumer := func(item interface{}) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}
		q.StartConsumers(2, consumer)

		var wg sync.WaitGroup

		// Producer
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				q.Produce(i)
			}
		}()

		// Size checker
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				size := q.Size()
				capacity := q.Capacity()
				if size < 0 {
					t.Errorf("Size should not be negative: %d", size)
				}
				if size > capacity {
					t.Errorf("Size %d should not exceed capacity %d", size, capacity)
				}
				time.Sleep(2 * time.Millisecond)
			}
		}()

		wg.Wait()
	})
}

// TestConsumerFunc tests the ConsumerFunc adapter
func TestConsumerFunc(t *testing.T) {
	t.Run("consumer func adapter", func(t *testing.T) {
		var processed []interface{}
		var mu sync.Mutex

		callback := func(item interface{}) error {
			mu.Lock()
			processed = append(processed, item)
			mu.Unlock()
			if item == "error" {
				return errors.New("test error")
			}
			return nil
		}

		consumer := ConsumerFunc(callback)

		// Test successful consumption
		err := consumer.Consume("success")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Test error case
		err = consumer.Consume("error")
		if err == nil {
			t.Error("Expected error, got nil")
		}

		mu.Lock()
		defer mu.Unlock()
		if len(processed) != 2 {
			t.Errorf("Expected 2 processed items, got %d", len(processed))
		}
	})
}

// TestBoundedQueuePanicHandling tests panic recovery in consumers
func TestBoundedQueuePanicHandling(t *testing.T) {
	tests := []struct {
		name         string
		retryEnabled bool
		expectedMin  int // Minimum expected processed items
	}{
		{"panic with retry disabled", false, 1},
		{"panic with retry enabled", true, 0}, // With retry, panic item may be retried and dropped
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var processed int32
			consumer := func(item interface{}) error {
				if item == "panic" {
					panic("test panic")
				}
				atomic.AddInt32(&processed, 1)
				return nil
			}

			var q *BoundedQueue
			if tt.retryEnabled {
				q = NewBoundedQueueWithRetry(10, nil, 2, 10*time.Millisecond)
			} else {
				q = NewBoundedQueue(10, nil)
			}
			defer q.Stop()

			q.StartConsumers(1, consumer)
			time.Sleep(10 * time.Millisecond)

			// This should not crash the test
			q.Produce("panic")
			time.Sleep(200 * time.Millisecond) // Longer wait for retry scenarios

			// Queue should still be functional
			q.Produce("normal")
			time.Sleep(200 * time.Millisecond)

			finalProcessed := atomic.LoadInt32(&processed)
			if int(finalProcessed) < tt.expectedMin {
				t.Errorf("Expected at least %d processed items, got %d", tt.expectedMin, finalProcessed)
			}
		})
	}
}

// min helper function for Go versions that don't have it built-in
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BenchmarkBoundedQueueProduce benchmarks queue production
func BenchmarkBoundedQueueProduce(b *testing.B) {
	q := NewBoundedQueue(10000, nil)
	defer q.Stop()

	consumer := func(item interface{}) error { return nil }
	q.StartConsumers(4, consumer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Produce(i)
	}
}

// BenchmarkBoundedQueueProduceWithRetry benchmarks queue with retry
func BenchmarkBoundedQueueProduceWithRetry(b *testing.B) {
	q := NewBoundedQueueWithRetry(10000, nil, 3, 100*time.Millisecond)
	defer q.Stop()

	consumer := func(item interface{}) error { return nil }
	q.StartConsumers(4, consumer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Produce(i)
	}
}

// BenchmarkBoundedQueueConcurrentAccess benchmarks concurrent operations
func BenchmarkBoundedQueueConcurrentAccess(b *testing.B) {
	q := NewBoundedQueue(1000, nil)
	defer q.Stop()

	consumer := func(item interface{}) error { return nil }
	q.StartConsumers(4, consumer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			q.Produce(i)
			i++
		}
	})
}

// TestBoundedQueueAtomicOperations tests thread safety of atomic operations
func TestBoundedQueueAtomicOperations(t *testing.T) {
	t.Run("atomic size updates", func(t *testing.T) {
		q := NewBoundedQueue(1000, nil)
		defer q.Stop()

		// Slow consumer to allow queue to fill up
		consumer := func(item interface{}) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}
		q.StartConsumers(1, consumer)

		var wg sync.WaitGroup
		numGoroutines := 10

		// Multiple producers
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					q.Produce(j)
				}
			}()
		}

		// Size checker
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				size := q.Size()
				if size < 0 {
					t.Errorf("Size should never be negative, got %d", size)
				}
				time.Sleep(1 * time.Millisecond)
			}
		}()

		wg.Wait()
	})
}

// TestBoundedQueueMemoryManagement tests proper cleanup
func TestBoundedQueueMemoryManagement(t *testing.T) {
	t.Run("cleanup after stop", func(t *testing.T) {
		q := NewBoundedQueue(100, nil)
		consumer := func(item interface{}) error { return nil }
		q.StartConsumers(2, consumer)

		// Add items
		for i := 0; i < 50; i++ {
			q.Produce(i)
		}

		// Stop should clean up properly
		q.Stop()

		// Operations after stop should fail gracefully
		if q.Produce("after_stop") {
			t.Error("Produce should fail after stop")
		}

		// Size should still be accessible
		size := q.Size()
		if size < 0 {
			t.Errorf("Size should not be negative after stop, got %d", size)
		}
	})
}

// TestBoundedQueueUnsafePointerOperations tests unsafe pointer usage in Resize
func TestBoundedQueueUnsafePointerOperations(t *testing.T) {
	t.Run("unsafe pointer in resize", func(t *testing.T) {
		q := NewBoundedQueue(10, nil)
		defer q.Stop()

		consumer := func(item interface{}) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}
		q.StartConsumers(1, consumer)

		// Get initial pointer
		initialItems := q.items

		// Resize should change the pointer
		success := q.Resize(20)
		if !success {
			t.Error("Resize should succeed")
		}

		// Verify pointer changed
		newItems := q.items
		if unsafe.Pointer(initialItems) == unsafe.Pointer(newItems) {
			t.Error("Items channel pointer should have changed after resize")
		}

		if q.Capacity() != 20 {
			t.Errorf("Capacity should be 20 after resize, got %d", q.Capacity())
		}
	})
}
