package gateway

import (
	gwService "github.com/mycontroller-org/server/v2/pkg/service/gateway"
	"github.com/mycontroller-org/server/v2/pkg/start/common"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"go.uber.org/zap"
)

// Init func
func Init(configFile string) {
	common.InitBasicServices(configFile, initServices, closeServices)
}

func initServices() {
	// init gateway listener
	err := gwService.Start(&store.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("failed to init gateway service listener", zap.Error(err))
	}
}

func closeServices() {
	// close gateway service
	gwService.Close()
}
