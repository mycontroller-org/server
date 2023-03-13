package msglogger

import (
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
)

// No Operation message logger struct
type NoopMessageLogger struct {
}

// Start implementation
func (vm *NoopMessageLogger) Start() {}

// Close implementation
func (vm *NoopMessageLogger) Close() {}

// AsyncWrite implementation
func (vm *NoopMessageLogger) AsyncWrite(rawMsg *msgTY.RawMessage) {}
