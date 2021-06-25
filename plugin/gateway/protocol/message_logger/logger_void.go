package msglogger

import (
	msgml "github.com/mycontroller-org/server/v2/pkg/model/message"
)

// VoidMessageLogger struct
type VoidMessageLogger struct {
}

// Start implementation
func (vm *VoidMessageLogger) Start() {}

// Close implementation
func (vm *VoidMessageLogger) Close() {}

// AsyncWrite implementation
func (vm *VoidMessageLogger) AsyncWrite(rawMsg *msgml.RawMessage) {}
