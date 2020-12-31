package allinone

import (
	msgPRO "github.com/mycontroller-org/backend/v2/pkg/processor/message"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	"github.com/mycontroller-org/backend/v2/pkg/service/init/common"
	"github.com/mycontroller-org/backend/v2/pkg/service/init/core"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	resourceSVC "github.com/mycontroller-org/backend/v2/pkg/service/resource"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	gwService "github.com/mycontroller-org/backend/v2/plugin/gateway/service"
	"go.uber.org/zap"
)

// Init func
func Init(handlerFunc func()) {
	common.InitBasicServices(wrapHandlerFunc(handlerFunc), closeServices)
}

func initServices() {
	stg.Init() // storage
	mts.Init() // metrics

	core.StartupJobs(&cfg.CFG.StartupJobs)
	core.UpdateInitialUser()

	// start message processing engine
	msgPRO.Init()

	// init resource server
	err := resourceSVC.Init()
	if err != nil {
		zap.L().Fatal("Failed to init resource service listener", zap.Error(err))
	}

	// init gateway listener
	err = gwService.Init(cfg.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("Failed to init gateway service listener", zap.Error(err))
	}
}

func wrapHandlerFunc(handlerFunc func()) func() {
	return func() {
		initServices()
		if handlerFunc != nil {
			go handlerFunc()
		}
	}
}

func closeServices() {
	// close gateway service
	gwService.Close()

	// close resource service
	resourceSVC.Close()

	// stop engine
	zap.L().Debug("Closing message process engine")
	msgPRO.Close()

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
