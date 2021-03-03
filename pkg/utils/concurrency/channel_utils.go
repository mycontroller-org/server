package concurrency

import "sync"

// Channel struct
type Channel struct {
	CH     chan interface{}
	closed bool
	mutex  sync.Mutex
}

// NewChannel returns new instance
func NewChannel(capacity int) *Channel {
	return &Channel{CH: make(chan interface{}, capacity)}
}

// SafeClose closes only once
func (c *Channel) SafeClose() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.closed {
		close(c.CH)
		c.closed = true
	}
}

// SafeSend sends data safely
func (c *Channel) SafeSend(data interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.closed {
		c.CH <- data
	}
}

// IsClosed returns status
func (c *Channel) IsClosed() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.closed
}
