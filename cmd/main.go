package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mycontroller-org/backend/cmd/app/handler"
	gwAPI "github.com/mycontroller-org/backend/pkg/api/gateway"
	msgPRO "github.com/mycontroller-org/backend/pkg/engine/message"
	"github.com/mycontroller-org/backend/pkg/mcbus"
	svc "github.com/mycontroller-org/backend/pkg/service"
)

func init() {
	preInitFn := func() {
		mcbus.Start()
	}
	start := time.Now()
	svc.Init(preInitFn, nil)
	zap.L().Debug("Init complete", zap.String("timeTaken", time.Since(start).String()))
}

func main() {
	defer zap.L().Sync()
	// call shutdown handler
	go handleShutdown()

	// start engine
	msgPRO.Init()

	// load gateways
	start := time.Now()
	gwAPI.LoadGateways()
	zap.L().Debug("Load gateways done.", zap.String("timeTaken", time.Since(start).String()))

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
