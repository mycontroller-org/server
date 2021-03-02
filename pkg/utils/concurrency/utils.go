package concurrency

import "sync"

// SafeBool is safe to use in concurrently.
type SafeBool struct {
	mutex sync.Mutex
	state bool
}

// IsSet returns state
func (sb *SafeBool) IsSet() bool {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	return sb.state
}

// Set updates the state to true
func (sb *SafeBool) Set() {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	sb.state = true
}

// Reset updates the state to false
func (sb *SafeBool) Reset() {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	sb.state = false
}
