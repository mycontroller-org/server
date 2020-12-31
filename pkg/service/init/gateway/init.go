package gateway

import (
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	"github.com/mycontroller-org/backend/v2/pkg/service/init/common"
	gwService "github.com/mycontroller-org/backend/v2/plugin/gateway/service"
	"go.uber.org/zap"
)

// Init func
func Init() {
	common.InitBasicServices(initServices, closeServices)
}

func initServices() {
	// init gateway listener
	err := gwService.Init(cfg.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("Failed to init gateway service listener", zap.Error(err))
	}
}

func closeServices() {
	// close gateway service
	gwService.Close()
}
