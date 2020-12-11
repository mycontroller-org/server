package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mycontroller-org/backend/v2/cmd/app/handler"
	gwAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	msgPRO "github.com/mycontroller-org/backend/v2/pkg/processor/message"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	"github.com/mycontroller-org/backend/v2/pkg/storage/export"
)

func init() {
	preInitFn := func() {
		mcbus.Start()
	}
	postInitFn := func(cfg *config.Config) {
		// call shutdown handler
		go handleShutdown()

		// startup jobs
		startupJobs(&cfg.StartupJobs)

		// start engine
		msgPRO.Init()

		// load gateways
		gwStart := time.Now()
		gwAPI.LoadGateways()
		zap.L().Debug("Load gateways done.", zap.String("timeTaken", time.Since(gwStart).String()))
	}

	start := time.Now()
	svc.Init(preInitFn, postInitFn)
	zap.L().Debug("Init complete", zap.String("timeTaken", time.Since(start).String()))
}

func startupJobs(cfg *config.Startup) {
	if cfg.Importer.Enabled {
		err := export.ExecuteImport(cfg.Importer.TargetDirectory, cfg.Importer.Type)
		if err != nil {
			zap.L().Error("Failed to load exported files", zap.Error(err))
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

	// unload gateways
	zap.L().Debug("Unloading gateways")
	gwAPI.UnloadGateways()

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
