package service

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	"github.com/mycontroller-org/backend/v2/pkg/export"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	msgPRO "github.com/mycontroller-org/backend/v2/pkg/processor/message"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	"github.com/mycontroller-org/backend/v2/pkg/service/logger"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	resourceSVC "github.com/mycontroller-org/backend/v2/pkg/service/resource"
	sch "github.com/mycontroller-org/backend/v2/pkg/service/scheduler"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	gwService "github.com/mycontroller-org/backend/v2/plugin/gateway/service"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

const (
	typeAllInOne = "all-in-one"
	typeGateway  = "gateway"
	typeCore     = "core"
)

var initType = ""

// CloseServices func
func CloseServices() {
	// close scheduler
	if sch.SVC != nil {
		sch.SVC.Close()
	}
	// Close storage and metric database
	if stg.SVC != nil {
		err := stg.SVC.Close()
		if err != nil {
			zap.L().Error("Failed to close storage database")
		}
	}
	if mts.SVC != nil {
		if mts.SVC != nil {
			err := mts.SVC.Close()
			if err != nil {
				zap.L().Error("Failed to close metrics database")
			}
		}
	}
}

// InitAllInOne func
func InitAllInOne() {
	initFunc(typeAllInOne)
}

// InitCore func
func InitCore() {
	initFunc(typeCore)
}

// InitGateway func
func InitGateway() {
	initFunc(typeGateway)
}

// init func
func initFunc(targetType string) {
	initType = targetType

	cfg.Init()
	logger.Init()

	mcbus.InitBus(cfg.CFG.Bus) // bus
	sch.Init()                 // scheduler

	if initType == typeAllInOne || initType == typeCore {
		stg.Init() // storage
		mts.Init() // metrics
	}

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

	if initType == typeAllInOne || initType == typeCore {
		startupJobs(&cfg.CFG.StartupJobs)

		// add default admin user
		updateInitialUser()

		// start message processing engine
		msgPRO.Init()

		// init resource server
		err = resourceSVC.Init()
		if err != nil {
			zap.L().Fatal("Failed to init resource service listener", zap.Error(err))
		}
	}

	if initType == typeAllInOne || initType == typeGateway {
		// init gateway listener
		err = gwService.Init(cfg.CFG.Gateway)
		if err != nil {
			zap.L().Fatal("Failed to init gateway service listener", zap.Error(err))
		}
	}
}

func startupJobs(cfg *config.Startup) {
	if cfg.Importer.Enabled {
		err := export.ExecuteImport(cfg.Importer.TargetDirectory, cfg.Importer.Type)
		if err != nil {
			zap.L().WithOptions(zap.AddCallerSkip(10)).Error("Failed to load exported files", zap.String("error", err.Error()))
		}
	}
}

func updateInitialUser() {
	pagination := &stgml.Pagination{
		Limit: 1,
	}
	users, err := userAPI.List(nil, pagination)
	if err != nil {
		zap.L().Error("failed to users", zap.Error(err))
	}
	if users.Count == 0 {
		adminUser := &userML.User{
			Username: "admin",
			Password: "admin",
			FullName: "Admin User",
			Email:    "admin@example.com",
		}
		err = userAPI.Save(adminUser)
		if err != nil {
			zap.L().Error("failed to create default admin user", zap.Error(err))
		}
	}
}

func handleShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// waiting for signal
	sig := <-sigs
	close(sigs)

	start := time.Now()
	zap.L().Info("Shutdown initiated..", zap.Any("signal", sig))

	if initType == typeAllInOne || initType == typeGateway {
		// close gateway service
		gwService.Close()
	}

	if initType == typeAllInOne || initType == typeCore {
		// close resource service
		resourceSVC.Close()
	}

	// stop engine
	zap.L().Debug("Closing message process engine")
	msgPRO.Close()

	// close services
	zap.L().Debug("Closing all other services")
	CloseServices()

	zap.L().Debug("Close services are done", zap.String("timeTaken", time.Since(start).String()))
	zap.L().Debug("Bye, See you soon :)")

	// stop web/api service
	os.Exit(0)
}
