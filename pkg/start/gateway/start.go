package gateway

import (
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	gwService "github.com/mycontroller-org/backend/v2/pkg/service/gateway"
	"github.com/mycontroller-org/backend/v2/pkg/start/common"
	"go.uber.org/zap"
)

// Init func
func Init() {
	common.InitBasicServices(initServices, closeServices)
}

func initServices() {
	// init gateway listener
	err := gwService.Start(&cfg.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("failed to init gateway service listener", zap.Error(err))
	}
}

func closeServices() {
	// close gateway service
	gwService.Close()
}
