package handler

import (
	handlerSVC "github.com/mycontroller-org/server/v2/pkg/service/handler"
	"github.com/mycontroller-org/server/v2/pkg/start/common"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"go.uber.org/zap"
)

// Init func
func Init(configFile string) {
	common.InitBasicServices(configFile, initServices, closeServices)
}

func initServices() {
	// init handler listener
	err := handlerSVC.Start(&store.CFG.Handler)
	if err != nil {
		zap.L().Fatal("failed to init handler service listener", zap.Error(err))
	}
}

func closeServices() {
	// close handler service
	handlerSVC.Close()
}
