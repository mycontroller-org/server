package core

import (
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	"github.com/mycontroller-org/backend/v2/pkg/export"
	"github.com/mycontroller-org/backend/v2/pkg/init/common"
	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	fwdplSVC "github.com/mycontroller-org/backend/v2/pkg/service/forward_payload"
	msgProcessor "github.com/mycontroller-org/backend/v2/pkg/service/message_processor"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	resourceSVC "github.com/mycontroller-org/backend/v2/pkg/service/resource"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// Init func
func Init(handlerFunc func()) {
	common.InitBasicServices(wrapHandlerFunc(handlerFunc), closeServices)
}

func initServices() {
	stg.Init() // storage
	mts.Init() // metrics

	StartupJobs(&cfg.CFG.StartupJobs)
	UpdateInitialUser()

	// start message processing engine
	msgProcessor.Init()

	// init resource server
	err := resourceSVC.Init()
	if err != nil {
		zap.L().Fatal("Error on init resource service listener", zap.Error(err))
	}

	// init payload forward service
	err = fwdplSVC.Init()
	if err != nil {
		zap.L().Fatal("Error on init forward payload service", zap.Error(err))
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
func StartupJobs(cfg *config.Startup) {
	if cfg.Importer.Enabled {
		err := export.ExecuteImport(cfg.Importer.TargetDirectory, cfg.Importer.Type)
		if err != nil {
			zap.L().WithOptions(zap.AddCallerSkip(10)).Error("Failed to load exported files", zap.String("error", err.Error()))
		}
	}
}

// UpdateInitialUser func
func UpdateInitialUser() {
	pagination := &stgml.Pagination{
		Limit: 1,
	}
	users, err := userAPI.List(nil, pagination)
	if err != nil {
		zap.L().Error("failed to users", zap.Error(err))
	}
	if users.Count == 0 {
		adminUser := &userML.User{
			Username: "admin",
			Password: "admin",
			FullName: "Admin User",
			Email:    "admin@example.com",
		}
		err = userAPI.Save(adminUser)
		if err != nil {
			zap.L().Error("failed to create default admin user", zap.Error(err))
		}
	}
}

func closeServices() {

	// close forward payload service
	fwdplSVC.Close()

	// close resource service
	resourceSVC.Close()

	// stop engine
	zap.L().Debug("Closing message process engine")
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
