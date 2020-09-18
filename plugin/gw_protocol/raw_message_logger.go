package gwprotocol

import (
	"fmt"
	"sync"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/util"
	"go.uber.org/zap"
)

// RawMessageLogger struct
type RawMessageLogger struct {
	Config         *gwml.Config                          // Gateway config
	MsgFormatterFn func(rawMsg *msgml.RawMessage) string // should supply a func to return parsed message
	stopCh         chan bool                             // this channel used to terminate the serial port listener
	msgQueue       []*msgml.RawMessage                   // Messages will be added in to this queue and dump into the file N seconds once
	mutex          sync.Mutex                            // lock to access the queue
}

// Start func
// inits the the queue and channel and starts the async runner
func (rml *RawMessageLogger) Start() {
	rml.mutex.Lock()
	rml.mutex.Unlock()
	rml.msgQueue = make([]*msgml.RawMessage, 0)
	rml.stopCh = make(chan bool)
	go util.AsyncRunner(rml.write, 1*time.Second, rml.stopCh)
}

// Close terminates the async runner
func (rml *RawMessageLogger) Close() {
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
func (rml *RawMessageLogger) AsyncWrite(rawMsg *msgml.RawMessage) {
	rml.mutex.Lock()
	defer rml.mutex.Unlock()
	rml.msgQueue = append(rml.msgQueue, rawMsg)
}

// write dumps the queue data into disk on a file
func (rml *RawMessageLogger) write() {
	rml.mutex.Lock()
	defer rml.mutex.Unlock()
	if len(rml.msgQueue) > 0 {
		// generate filename to store log data
		logDirectory := model.DirectoryFullPath(model.DirectoryGatewayRawLogs)
		logFilename := fmt.Sprintf("gw_%s.log", rml.Config.ID)
		for _, rawMsg := range rml.msgQueue {
			msgStr := rml.MsgFormatterFn(rawMsg)
			err := util.AppendFile(logDirectory, logFilename, []byte(msgStr))
			if err != nil {
				zap.L().Error("Failed to write", zap.Error(err))
			}
		}
	}
	rml.msgQueue = nil
}
