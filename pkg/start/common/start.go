package common

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	sch "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/pkg/service/logger"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

// InitBasicServices func
func InitBasicServices(initCustomServices func(), closeCustomServices func()) {
	// 	defer func() {
	// 		err := zap.L().Sync()
	// 		if err != nil {
	// 			zap.L().Error("error on sync", zap.Error(err))
	// 		}
	// 	}()

	start := time.Now()

	// load configuration
	cfg, err := cfg.Load()
	if err != nil {
		zap.L().Fatal("failed to load configuration", zap.Error(err))
		return
	}
	store.InitConfig(cfg) // add config in to the common store

	logger.Load(cfg.Logger, "the MyController world")

	mcbus.Start(cfg.Bus) // bus
	sch.Init()           // scheduler

	// load root directories
	types.UpdateDirecotries(cfg.Directories)
	// create root directories
	err = utils.CreateDir(types.GetDirectoryDataRoot())
	if err != nil {
		zap.L().Fatal("failed to create root directory", zap.Error(err))
	}
	err = utils.CreateDir(types.GetDirectoryLogsRoot())
	if err != nil {
		zap.L().Fatal("failed to create root directory", zap.Error(err))
	}

	if initCustomServices != nil {
		initCustomServices()
	}

	zap.L().Debug("init complete", zap.String("timeTaken", time.Since(start).String()))

	// call handle shutdown
	handleShutdownEvent(closeCustomServices)
	handleShutdownSignal(closeCustomServices)
}

// handleShutdownSignal func
func handleShutdownSignal(closeCustomServices func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// waiting for signal
	sig := <-sigs
	close(sigs)

	zap.L().Info("shutdown initiated..", zap.Any("signal", sig))
	triggerShutdown(closeCustomServices)
}

// handleShutdownEvent func
func handleShutdownEvent(closeCustomServices func()) {
	shutdownFunc := func(data *busTY.BusData) {
		zap.L().Info("shutdown initiated..", zap.Any("signal", "internal event"))
		triggerShutdown(closeCustomServices)
	}
	_, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicInternalShutdown), shutdownFunc)
	if err != nil {
		zap.L().Fatal("error on subscribing shutdown event", zap.Error(err))
		return
	}

}

func triggerShutdown(closeCustomServices func()) {
	start := time.Now()

	// close other services
	if closeCustomServices != nil {
		closeCustomServices()
	}

	if sch.SVC != nil {
		sch.SVC.Close()
	}

	mcbus.Close()

	zap.L().Debug("closing services are done", zap.String("timeTaken", time.Since(start).String()))
	zap.L().Debug("bye, see you soon :)")

	// stop web/api service
	os.Exit(0)
}
