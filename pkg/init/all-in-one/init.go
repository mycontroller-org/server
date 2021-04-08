package allinone

import (
	"github.com/mycontroller-org/backend/v2/pkg/init/common"
	"github.com/mycontroller-org/backend/v2/pkg/init/core"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	fwdplSVC "github.com/mycontroller-org/backend/v2/pkg/service/forward_payload"
	gwService "github.com/mycontroller-org/backend/v2/pkg/service/gateway"
	handlerSVC "github.com/mycontroller-org/backend/v2/pkg/service/handlers"
	msgProcessor "github.com/mycontroller-org/backend/v2/pkg/service/message_processor"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	resourceSVC "github.com/mycontroller-org/backend/v2/pkg/service/resource"
	schedulerSVC "github.com/mycontroller-org/backend/v2/pkg/service/scheduler"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	taskSVC "github.com/mycontroller-org/backend/v2/pkg/service/task"
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
	core.StartupJobsExtra()

	// start message processing engine
	err := msgProcessor.Init()
	if err != nil {
		zap.L().Fatal("error on init message process service", zap.Error(err))
	}

	// init resource server
	err = resourceSVC.Init()
	if err != nil {
		zap.L().Fatal("error on init resource servicelistener", zap.Error(err))
	}

	// load notify handlers
	err = handlerSVC.Init(cfg.CFG.Handler)
	if err != nil {
		zap.L().Fatal("error on start notify handler service", zap.Error(err))
	}

	// init task engine
	err = taskSVC.Init(cfg.CFG.Task)
	if err != nil {
		zap.L().Fatal("error on init task engine service", zap.Error(err))
	}

	// init scheduler engine
	err = schedulerSVC.Init(cfg.CFG.Task)
	if err != nil {
		zap.L().Fatal("error on init scheduler service", zap.Error(err))
	}

	// init payload forward service
	err = fwdplSVC.Init()
	if err != nil {
		zap.L().Fatal("error on init forward payload service", zap.Error(err))
	}

	// init gateway listener
	err = gwService.Init(cfg.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("error on init gateway service listener", zap.Error(err))
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
	// close forward payload service
	fwdplSVC.Close()

	// close gateway service
	gwService.Close()

	// close resource service
	resourceSVC.Close()

	// close task service
	taskSVC.Close()

	// close scheduler service
	schedulerSVC.Close()

	// close notify handler service
	handlerSVC.Close()

	// stop engine
	msgProcessor.Close()

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
