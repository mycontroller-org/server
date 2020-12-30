package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mycontroller-org/backend/v2/cmd/app/handler"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	"github.com/mycontroller-org/backend/v2/pkg/export"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	msgPRO "github.com/mycontroller-org/backend/v2/pkg/processor/message"
	resourceSVC "github.com/mycontroller-org/backend/v2/pkg/resource_service"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	gwService "github.com/mycontroller-org/backend/v2/plugin/gateway/service"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func init() {
	start := time.Now()
	svc.Init(preInitFunc, postInitFunc)
	zap.L().Debug("Init complete", zap.String("timeTaken", time.Since(start).String()))
}

func preInitFunc(cfg *config.Config) {
	// init bus
	err := mcbus.InitBus(cfg.Bus)
	if err != nil {
		zap.L().Fatal("Failed to init bus client", zap.Error(err))
	}
}

func postInitFunc(cfg *config.Config) {
	// call shutdown handler
	go handleShutdown()

	// load root directories
	model.UpdateDirecotries(cfg.Directories)
	// create root directories
	err := utils.CreateDir(model.GetDirectoryDataRoot())
	if err != nil {
		zap.L().Fatal("Failed to create root directory", zap.Error(err))
	}
	err = utils.CreateDir(model.GetDirectoryLogsRoot())
	if err != nil {
		zap.L().Fatal("Failed to create root directory", zap.Error(err))
	}

	// startup jobs
	startupJobs(&cfg.StartupJobs)

	// add default admin user
	updateInitialUser()

	// start engine
	msgPRO.Init()

	// init resource server
	err = resourceSVC.Init()
	if err != nil {
		zap.L().Fatal("Failed to init resource service listener", zap.Error(err))
	}
	// init gateway listener
	err = gwService.Init(cfg.Gateway)
	if err != nil {
		zap.L().Fatal("Failed to init gateway service listener", zap.Error(err))
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

func startupJobs(cfg *config.Startup) {
	if cfg.Importer.Enabled {
		err := export.ExecuteImport(cfg.Importer.TargetDirectory, cfg.Importer.Type)
		if err != nil {
			zap.L().WithOptions(zap.AddCallerSkip(10)).Error("Failed to load exported files", zap.String("error", err.Error()))
		}
	}
}

func main() {
	defer zap.L().Sync()

	err := handler.StartHandler()
	if err != nil {
		zap.L().Fatal("Error on starting http handler", zap.Error(err))
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

	// close gateway service
	gwService.Close()

	// close resource service
	resourceSVC.Close()

	// stop engine
	zap.L().Debug("Closing message process engine")
	msgPRO.Close()

	// close services
	zap.L().Debug("Closing all other services")
	err := svc.Close()
	if err != nil {
		zap.L().Fatal("Error on closing services", zap.Error(err))
	}
	zap.L().Debug("Close services are done", zap.String("timeTaken", time.Since(start).String()))
	zap.L().Debug("Bye, See you soon :)")

	// stop web/api service
	os.Exit(0)
}
