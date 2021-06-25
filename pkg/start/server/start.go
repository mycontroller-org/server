package server

import (
	backupAPI "github.com/mycontroller-org/server/v2/pkg/backup"
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	fwdplSVC "github.com/mycontroller-org/server/v2/pkg/service/forward_payload"
	gwService "github.com/mycontroller-org/server/v2/pkg/service/gateway"
	gwMsgProcessor "github.com/mycontroller-org/server/v2/pkg/service/gateway_msg_processor"
	handlerSVC "github.com/mycontroller-org/server/v2/pkg/service/handler"
	metricsSVC "github.com/mycontroller-org/server/v2/pkg/service/metrics"
	resourceSVC "github.com/mycontroller-org/server/v2/pkg/service/resource"
	scheduleSVC "github.com/mycontroller-org/server/v2/pkg/service/schedule"
	storageSVC "github.com/mycontroller-org/server/v2/pkg/service/storage"
	taskSVC "github.com/mycontroller-org/server/v2/pkg/service/task"
	"github.com/mycontroller-org/server/v2/pkg/start/common"
	"go.uber.org/zap"
)

// Start func
func Start(handlerFunc func()) {
	common.InitBasicServices(wrapHandlerFunc(handlerFunc), closeServices)
}

func initServices() {
	storageSVC.Init(backupAPI.ExecuteImportStorage) // storage
	metricsSVC.Init()                               // metrics

	StartupJobs()
	StartupJobsExtra()

	// start message processing engine
	err := gwMsgProcessor.Start()
	if err != nil {
		zap.L().Fatal("error on init message process service", zap.Error(err))
	}

	// init resource server
	err = resourceSVC.Start()
	if err != nil {
		zap.L().Fatal("error on init resource servicelistener", zap.Error(err))
	}

	// load notify handlers
	err = handlerSVC.Start(&cfg.CFG.Handler)
	if err != nil {
		zap.L().Fatal("error on start notify handler service", zap.Error(err))
	}

	// init task engine
	err = taskSVC.Start(&cfg.CFG.Task)
	if err != nil {
		zap.L().Fatal("error on init task engine service", zap.Error(err))
	}

	// init scheduler engine
	err = scheduleSVC.Start(&cfg.CFG.Task)
	if err != nil {
		zap.L().Fatal("error on init scheduler service", zap.Error(err))
	}

	// init payload forward service
	err = fwdplSVC.Start()
	if err != nil {
		zap.L().Fatal("error on init forward payload service", zap.Error(err))
	}

	// start gateway listener
	err = gwService.Start(&cfg.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("error on starting gateway service listener", zap.Error(err))
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

// StartupJobs func
func StartupJobs() {
	RunSystemStartJobs()
}

func closeServices() {
	// close forward payload service
	fwdplSVC.Close()

	// close gateway service
	gwService.Close()

	// close task service
	taskSVC.Close()

	// close scheduler service
	scheduleSVC.Close()

	// close notify handler service
	handlerSVC.Close()

	// stop engine
	gwMsgProcessor.Close()

	// close resource service
	resourceSVC.Close()

	// Close storage and metric database
	if storageSVC.SVC != nil {
		err := storageSVC.SVC.Close()
		if err != nil {
			zap.L().Error("failed to close storage database")
		}
	}
	if metricsSVC.SVC != nil {
		if metricsSVC.SVC != nil {
			err := metricsSVC.SVC.Close()
			if err != nil {
				zap.L().Error("failed to close metrics database")
			}
		}
	}
}
