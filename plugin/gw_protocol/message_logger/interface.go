package msglogger

import (
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

// MessageLogger interface
type MessageLogger interface {
	Start()                              // Start the message logger
	Close()                              // Stop and close the message logger
	AsyncWrite(rawMsg *msgml.RawMessage) // write a message
}

// logger types
const (
	TypeVoidLogger = "void_logger"
	TypeFileLogger = "file_logger"
)
