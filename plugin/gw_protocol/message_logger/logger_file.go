package msglogger

import (
	"fmt"
	"sync"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	utils "github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

// FileMessageLogger struct
type FileMessageLogger struct {
	GatewayID        string                                // Gateway id
	MsgFormatterFunc func(rawMsg *msgml.RawMessage) string // should supply a func to return parsed message
	stopCh           chan bool                             // this channel used to terminate the serial port listener
	msgQueue         []*msgml.RawMessage                   // Messages will be added in to this queue and dump into the file N seconds once
	mutex            sync.Mutex                            // lock to access the queue
	Config           fileMessageLoggerConfig               // self configurations
}

// fileMessageLoggerConfig definition
type fileMessageLoggerConfig struct {
	Type          string // type of the message logger
	FlushInterval string // flush interval, how long once data to be dumped into disk
}

const (
	defaultFlushInterval = 1 * time.Second
	fileNamePrefix       = "gateway_logs"
)

// InitFileMessageLogger file logger
func InitFileMessageLogger(gatewayID string, config cmap.CustomMap, formatterFunc func(rawMsg *msgml.RawMessage) string) (*FileMessageLogger, error) {
	cfg := fileMessageLoggerConfig{}
	err := utils.MapToStruct(utils.TagNameNone, config, &cfg)
	if err != nil {
		return nil, err
	}
	fileLogger := &FileMessageLogger{
		GatewayID:        gatewayID,
		MsgFormatterFunc: formatterFunc,
		stopCh:           make(chan bool),
		msgQueue:         make([]*msgml.RawMessage, 0),
		Config:           cfg,
	}
	return fileLogger, nil
}

// Start func
// inits the the queue and channel and starts the async runner
func (rml *FileMessageLogger) Start() {
	rml.mutex.Lock()
	rml.mutex.Unlock()
	flushInterval, err := time.ParseDuration(rml.Config.FlushInterval)
	if err != nil {
		flushInterval = defaultFlushInterval
	}
	go utils.AsyncRunner(rml.write, flushInterval, rml.stopCh)
}

// Close terminates the async runner
func (rml *FileMessageLogger) Close() {
	rml.mutex.Lock()
	defer rml.mutex.Unlock()
	if rml.stopCh != nil {
		rml.stopCh <- true
		close(rml.stopCh)
		rml.stopCh = nil
	}
}

// AsyncWrite func
// adds the message into the queue and returns immediately
func (rml *FileMessageLogger) AsyncWrite(rawMsg *msgml.RawMessage) {
	cloned := rawMsg.Clone()
	cloned.Timestamp = time.Now()

	rml.mutex.Lock()
	defer rml.mutex.Unlock()
	rml.msgQueue = append(rml.msgQueue, cloned)
}

// write dumps the queue data into disk on a file
func (rml *FileMessageLogger) write() {
	rml.mutex.Lock()
	defer rml.mutex.Unlock()
	if len(rml.msgQueue) > 0 {
		// generate filename to store log data
		logFilename := fmt.Sprintf("%s_%s.log", fileNamePrefix, rml.GatewayID)
		for _, rawMsg := range rml.msgQueue {
			msgStr := rml.MsgFormatterFunc(rawMsg)
			err := utils.AppendFile(model.GetDirectoryGatewayLog(), logFilename, []byte(msgStr))
			if err != nil {
				zap.L().Error("Failed to write", zap.Error(err))
			}
		}
	}
	rml.msgQueue = nil
}
