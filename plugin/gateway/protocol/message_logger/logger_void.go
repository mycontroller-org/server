package msglogger

import (
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
)

// VoidMessageLogger struct
type VoidMessageLogger struct {
}

// Start implementation
func (vm *VoidMessageLogger) Start() {}

// Close implementation
func (vm *VoidMessageLogger) Close() {}

// AsyncWrite implementation
func (vm *VoidMessageLogger) AsyncWrite(rawMsg *msgTY.RawMessage) {}
