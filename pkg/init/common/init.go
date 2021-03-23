package common

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	sch "github.com/mycontroller-org/backend/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/service/logger"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

// InitBasicServices func
func InitBasicServices(initCustomServices, closeCustomServices func()) {
	defer func() {
		err := zap.L().Sync()
		if err != nil {
			zap.L().Error("error on sync", zap.Error(err))
		}
	}()

	start := time.Now()
	cfg.Init()
	logger.Init()

	mcbus.InitBus(cfg.CFG.Bus) // bus
	sch.Init()                 // scheduler

	// load root directories
	model.UpdateDirecotries(cfg.CFG.Directories)
	// create root directories
	err := utils.CreateDir(model.GetDirectoryDataRoot())
	if err != nil {
		zap.L().Fatal("Failed to create root directory", zap.Error(err))
	}
	err = utils.CreateDir(model.GetDirectoryLogsRoot())
	if err != nil {
		zap.L().Fatal("Failed to create root directory", zap.Error(err))
	}

	if initCustomServices != nil {
		initCustomServices()
	}

	zap.L().Debug("Init complete", zap.String("timeTaken", time.Since(start).String()))

	// call handle shutdown
	handleShutdown(closeCustomServices)
}

// handleShutdown func
func handleShutdown(closeCustomServices func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// waiting for signal
	sig := <-sigs
	close(sigs)

	start := time.Now()
	zap.L().Info("Shutdown initiated..", zap.Any("signal", sig))

	// close other services
	if closeCustomServices != nil {
		closeCustomServices()
	}

	if sch.SVC != nil {
		sch.SVC.Close()
	}

	mcbus.Close()

	zap.L().Debug("Close services are done", zap.String("timeTaken", time.Since(start).String()))
	zap.L().Debug("Bye, See you soon :)")

	// stop web/api service
	os.Exit(0)
}
