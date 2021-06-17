package handler

import (
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	handlerSVC "github.com/mycontroller-org/backend/v2/pkg/service/handler"
	"github.com/mycontroller-org/backend/v2/pkg/start/common"
	"go.uber.org/zap"
)

// Init func
func Init() {
	common.InitBasicServices(initServices, closeServices)
}

func initServices() {
	// init handler listener
	err := handlerSVC.Start(&cfg.CFG.Handler)
	if err != nil {
		zap.L().Fatal("failed to init handler service listener", zap.Error(err))
	}
}

func closeServices() {
	// close handler service
	handlerSVC.Close()
}
