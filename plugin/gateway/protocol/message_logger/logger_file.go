package msglogger

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	utils "github.com/mycontroller-org/backend/v2/pkg/utils"
	concurrencyUtils "github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// FileMessageLogger struct
type FileMessageLogger struct {
	GatewayID        string                                // Gateway id
	MsgFormatterFunc func(rawMsg *msgml.RawMessage) string // should supply a func to return parsed message
	runnerFlushLog   *concurrencyUtils.Runner              // this runner used to call flush log func
	runnerRotateLog  *concurrencyUtils.Runner              // this runner used to call rotate log func
	msgQueue         []*msgml.RawMessage                   // Messages will be added in to this queue and dump into the file N seconds once
	mutex            sync.Mutex                            // lock to access the queue
	Config           fileMessageLoggerConfig               // self configurations
	maxSize          int64                                 // maximum size of the file in bytes
	maxAge           time.Duration                         // maximum age of the file in duration
	maxBackup        int                                   // maximum number of backups, 0 or negative - no limit
	isRunning        bool                                  // start called
}

// fileMessageLoggerConfig definition
type fileMessageLoggerConfig struct {
	Type              string // type of the message logger
	FlushInterval     string // flush interval, how long once data to be dumped into disk
	LogRotateInterval string // how long once log rotate utils would be triggered
	MaxSize           string
	MaxAge            string
	MaxBackup         int
}

const (
	filenamePrefix                 = "gw"
	filenameFormatBackup           = "20060106_150405"
	defaultLogRotateWorkerInterval = time.Minute * 10
	defaultFlushWorkerInterval     = time.Second * 1
	defaultMaxSize                 = utils.MiB * 2
	defaultMaxAge                  = time.Hour * 24 * 3 // 3 days
	defaultMaxBackup               = 5
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

	if rml.isRunning {
		zap.L().Warn("this instance is in running state, close it and re initialize then start")
		return
	}
	rml.isRunning = true

	// update values
	rml.maxSize = utils.ParseSizeWithDefault(rml.Config.MaxSize, defaultMaxSize)
	rml.maxAge = utils.ToDuration(rml.Config.MaxAge, defaultMaxAge)
	rml.maxBackup = rml.Config.MaxBackup

	flushWorkerInterval := utils.ToDuration(rml.Config.FlushInterval, defaultFlushWorkerInterval)
	logRotateWorkerInterval := utils.ToDuration(rml.Config.LogRotateInterval, defaultLogRotateWorkerInterval)

	zap.L().Debug("starting message logger", zap.Int64("maxSize", rml.maxSize), zap.String("maxAge", rml.maxAge.String()),
		zap.Int("maxBackup", rml.maxBackup), zap.Duration("flushInterval", flushWorkerInterval),
		zap.Duration("logRotateInterval", logRotateWorkerInterval), zap.String("gateway", rml.GatewayID))

	rml.runnerFlushLog = concurrencyUtils.GetAsyncRunner(rml.workerFlushLog, flushWorkerInterval, false)
	rml.runnerRotateLog = concurrencyUtils.GetAsyncRunner(rml.workerRotateLog, logRotateWorkerInterval, false)

	go rml.runnerFlushLog.Start()
	go rml.runnerRotateLog.Start()
}

// Close terminates the async runner
func (rml *FileMessageLogger) Close() {
	rml.mutex.Lock()
	defer rml.mutex.Unlock()

	if rml.runnerFlushLog != nil {
		rml.runnerFlushLog.Close()
	}

	if rml.runnerRotateLog != nil {
		rml.runnerRotateLog.Close()
	}
}

func stopWorker(workerStopChan chan bool) {
	if workerStopChan != nil {
		workerStopChan <- true
		close(workerStopChan)
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

// workerFlushLog dumps the queue data into disk on a file
func (rml *FileMessageLogger) workerFlushLog() {
	rml.mutex.Lock()
	defer rml.mutex.Unlock()
	if len(rml.msgQueue) > 0 {
		for _, rawMsg := range rml.msgQueue {
			msgStr := rml.MsgFormatterFunc(rawMsg)
			err := utils.AppendFile(model.GetDirectoryGatewayLog(), rml.getFilename(), []byte(msgStr))
			if err != nil {
				zap.L().Error("Failed to write", zap.Error(err), zap.String("gateway", rml.GatewayID))
			}
		}
	}
	rml.msgQueue = nil
}

func (rml *FileMessageLogger) workerRotateLog() {
	rml.mutex.Lock()
	defer rml.mutex.Unlock()

	files, err := utils.ListFiles(model.GetDirectoryGatewayLog())
	if err != nil {
		zap.L().Error("Failed to get log files", zap.Error(err), zap.String("gateway", rml.GatewayID), zap.String("directory", model.GetDirectoryGatewayLog()))
		return
	}

	liveFilename := rml.getFilename()

	// check live file size
	for _, file := range files {
		if file.Name == liveFilename {
			if file.Size >= rml.maxSize {
				newFilenameFull := fmt.Sprintf("%s/%s.%s", model.GetDirectoryGatewayLog(), liveFilename, time.Now().Format(filenameFormatBackup))
				liveFilenameFull := fmt.Sprintf("%s/%s", model.GetDirectoryGatewayLog(), liveFilename)
				zap.L().Debug("Renaming file", zap.Any("size", file.Size), zap.Any("new name", newFilenameFull))
				err = os.Rename(liveFilenameFull, newFilenameFull)
				if err != nil {
					zap.L().Error("Failed to rename log file", zap.Error(err), zap.String("gateway", rml.GatewayID), zap.String("currentPath", liveFilenameFull), zap.String("newPath", newFilenameFull))
				}
			}
			break
		}
	}

	// check max age
	maxAgeTime := time.Now().Add(-1 * rml.maxAge)
	for _, file := range files {
		if strings.HasPrefix(file.Name, liveFilename+".") && file.ModifiedTime.Before(maxAgeTime) {
			filenameFull := fmt.Sprintf("%s/%s", model.GetDirectoryGatewayLog(), file.Name)
			zap.L().Debug("Files for deletion, max age", zap.Any("filename", file.Name))
			err = os.Remove(filenameFull)
			if err != nil {
				zap.L().Error("Failed to remove log file", zap.Error(err), zap.String("gateway", rml.GatewayID), zap.String("filename", filenameFull))
			}
		}
	}

	// check maximum backup
	// get file names
	if rml.maxBackup > 0 {
		filenames := []string{}
		backupFilesCount := 0
		for _, file := range files {
			if strings.HasPrefix(file.Name, liveFilename+".") {
				filenames = append(filenames, file.Name)
				backupFilesCount++
			}
		}
		sort.Strings(filenames)
		if backupFilesCount > rml.maxBackup {
			deletionFilenames := filenames[:len(filenames)-rml.maxBackup]
			zap.L().Debug("Log files for deletion", zap.Any("all", filenames), zap.Any("deletion", deletionFilenames))
			for _, filename := range deletionFilenames {
				filenameFull := fmt.Sprintf("%s/%s", model.GetDirectoryGatewayLog(), filename)
				err = os.Remove(filenameFull)
				if err != nil {
					zap.L().Error("Failed to delete log file", zap.Error(err), zap.String("gateway", rml.GatewayID), zap.String("filename", filenameFull))
				}
			}
		}
	}
}

func (rml *FileMessageLogger) getFilename() string {
	return fmt.Sprintf("%s_%s.log", filenamePrefix, rml.GatewayID)
}
