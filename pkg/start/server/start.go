package server

import (
	backupAPI "github.com/mycontroller-org/server/v2/pkg/backup"
	bkpMap "github.com/mycontroller-org/server/v2/pkg/backup/bkp_map"
	metricSVC "github.com/mycontroller-org/server/v2/pkg/service/database/metric"
	storageSVC "github.com/mycontroller-org/server/v2/pkg/service/database/storage"
	deletionSVC "github.com/mycontroller-org/server/v2/pkg/service/deletion"
	fwdplSVC "github.com/mycontroller-org/server/v2/pkg/service/forward_payload"
	gwService "github.com/mycontroller-org/server/v2/pkg/service/gateway"
	gwMsgProcessor "github.com/mycontroller-org/server/v2/pkg/service/gateway_msg_processor"
	handlerSVC "github.com/mycontroller-org/server/v2/pkg/service/handler"
	resourceSVC "github.com/mycontroller-org/server/v2/pkg/service/resource"
	scheduleSVC "github.com/mycontroller-org/server/v2/pkg/service/schedule"
	systemJobs "github.com/mycontroller-org/server/v2/pkg/service/system_jobs"
	taskSVC "github.com/mycontroller-org/server/v2/pkg/service/task"
	vaService "github.com/mycontroller-org/server/v2/pkg/service/virtual_assistant"
	"github.com/mycontroller-org/server/v2/pkg/start/common"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"go.uber.org/zap"
)

// Start func
func Start(configFile string, handlerFunc func()) {
	common.InitBasicServices(configFile, wrapHandlerFunc(handlerFunc), closeServices)
}

func startServices() {
	stgSVC, err := storageSVC.Init(store.CFG.Database.Storage, store.CFG.Logger) // storage
	if err != nil {
		zap.L().Fatal("error on init storage database", zap.Error(err))
	}
	store.InitStorage(stgSVC) // load storage database client
	err = storageSVC.RunImport(store.STORAGE, bkpMap.ImportMap, backupAPI.ExecuteImportStorage)
	if err != nil {
		zap.L().Fatal("error on import", zap.Error(err))
	}

	mtgSVC, err := metricSVC.Init(store.CFG.Database.Metric, store.CFG.Logger) // metric
	if err != nil {
		zap.L().Fatal("error on init metric database", zap.Error(err))
	}
	store.InitMetric(mtgSVC) // load storage database client

	StartupJobs()
	StartupJobsExtra()

	// start message processing engine
	err = gwMsgProcessor.Start()
	if err != nil {
		zap.L().Fatal("error on init message process service", zap.Error(err))
	}

	// init resource server
	err = resourceSVC.Start()
	if err != nil {
		zap.L().Fatal("error on init resource servicelistener", zap.Error(err))
	}

	// load notify handlers
	err = handlerSVC.Start(&store.CFG.Handler)
	if err != nil {
		zap.L().Fatal("error on start notify handler service", zap.Error(err))
	}

	// init task engine
	err = taskSVC.Start(&store.CFG.Task)
	if err != nil {
		zap.L().Fatal("error on init task engine service", zap.Error(err))
	}

	// init scheduler engine
	err = scheduleSVC.Start(&store.CFG.Task)
	if err != nil {
		zap.L().Fatal("error on init scheduler service", zap.Error(err))
	}

	// init payload forward service
	err = fwdplSVC.Start()
	if err != nil {
		zap.L().Fatal("error on init forward payload service", zap.Error(err))
	}

	// start gateway listener
	err = gwService.Start(&store.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("error on starting gateway service listener", zap.Error(err))
	}

	// start virtual assistant listener
	err = vaService.Start(&store.CFG.VirtualAssistant)
	if err != nil {
		zap.L().Fatal("error on starting virtual assistant service listener", zap.Error(err))
	}

	// start system jobs listener
	err = systemJobs.Start()
	if err != nil {
		zap.L().Fatal("error on starting system jobs service listener", zap.Error(err))
	}

	// start deletion service
	err = deletionSVC.Start()
	if err != nil {
		zap.L().Fatal("error on starting deletion service", zap.Error(err))
	}
}

func wrapHandlerFunc(handlerFunc func()) func() {
	return func() {
		startServices()
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
	// close deletion service
	err := deletionSVC.Close()
	if err != nil {
		zap.L().Fatal("error on closgin deletion service", zap.Error(err))
	}

	// close forward payload service
	fwdplSVC.Close()

	// close gateway service
	gwService.Close()

	// close task service
	taskSVC.Close()

	// close system jobs service
	systemJobs.Close()

	// close scheduler service
	scheduleSVC.Close()

	// close notify handler service
	handlerSVC.Close()

	// stop engine
	gwMsgProcessor.Close()

	// close resource service
	resourceSVC.Close()

	// Close storage and metric database
	if store.STORAGE != nil {
		err := store.STORAGE.Close()
		if err != nil {
			zap.L().Error("failed to close storage database")
		}
	}
	if store.METRIC != nil {
		err := store.METRIC.Close()
		if err != nil {
			zap.L().Error("failed to close metrics database")
		}
	}

}
