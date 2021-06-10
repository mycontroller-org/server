package handler

import (
	"github.com/mycontroller-org/backend/v2/pkg/init/common"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	handlerSVC "github.com/mycontroller-org/backend/v2/pkg/service/handler"
	"go.uber.org/zap"
)

// Init func
func Init() {
	common.InitBasicServices(initServices, closeServices)
}

func initServices() {
	// init handler listener
	err := handlerSVC.Init(cfg.CFG.Handler)
	if err != nil {
		zap.L().Fatal("failed to init handler service listener", zap.Error(err))
	}
}

func closeServices() {
	// close handler service
	handlerSVC.Close()
}
