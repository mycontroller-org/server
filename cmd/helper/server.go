package helper

import (
	"context"
	"time"

	"github.com/gorilla/mux"
	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	"github.com/mycontroller-org/server/v2/pkg/configuration"
	"github.com/mycontroller-org/server/v2/pkg/encryption"
	httpRouter "github.com/mycontroller-org/server/v2/pkg/http_router"
	deletionSVC "github.com/mycontroller-org/server/v2/pkg/service/deletion"
	fwdPayloadSVC "github.com/mycontroller-org/server/v2/pkg/service/forward_payload"
	gatewaySVC "github.com/mycontroller-org/server/v2/pkg/service/gateway"
	gwMsgProcessorSVC "github.com/mycontroller-org/server/v2/pkg/service/gateway_msg_processor"
	handlerSVC "github.com/mycontroller-org/server/v2/pkg/service/handler"
	httpListenerSVC "github.com/mycontroller-org/server/v2/pkg/service/http_listener"
	resourceSVC "github.com/mycontroller-org/server/v2/pkg/service/resource"
	schedulerSVC "github.com/mycontroller-org/server/v2/pkg/service/scheduler"
	systemJobsSVC "github.com/mycontroller-org/server/v2/pkg/service/system_jobs"
	taskSVC "github.com/mycontroller-org/server/v2/pkg/service/task"
	vaSVC "github.com/mycontroller-org/server/v2/pkg/service/virtual_assistant"
	websocketSVC "github.com/mycontroller-org/server/v2/pkg/service/websocket"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	"github.com/mycontroller-org/server/v2/pkg/upgrade"
	templateUtils "github.com/mycontroller-org/server/v2/pkg/utils/template"
	variablesUtils "github.com/mycontroller-org/server/v2/pkg/utils/variables"
	"github.com/mycontroller-org/server/v2/pkg/version"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	ctx     context.Context
	config  *config.Config
	logger  *zap.Logger
	enc     *encryption.Encryption
	metric  metricTY.Plugin
	storage storageTY.Plugin
	bus     busTY.Plugin
	api     *entitiesAPI.API

	// services
	coreSchedulerSVC    schedulerTY.CoreScheduler // core scheduler, used to execute all the cron jobs
	resourceSVC         serviceTY.Service
	messageProcessorSVC serviceTY.Service
	taskSVC             serviceTY.Service
	schedulerSVC        serviceTY.Service
	deletionSVC         serviceTY.Service
	fwdPayloadSVC       serviceTY.Service
	gatewaySVC          serviceTY.Service
	handlerSVC          serviceTY.Service
	systemJobsSVC       serviceTY.Service
	virtualAssistantSVC serviceTY.Service
	websocketSVC        serviceTY.Service
}

func (s *Server) Start(ctx context.Context, configFilePath string) error {
	s.ctx = ctx
	router := mux.NewRouter()

	// load config
	cfg, err := configuration.Get(configFilePath)
	if err != nil {
		return err
	}

	// load logger
	ctx, logger := loadLogger(ctx, cfg.Logger, "server")

	// set environment values
	err = setEnvironmentVariables(cfg)
	if err != nil {
		logger.Error("error on setting environment variables", zap.Error(err))
		return err
	}

	// load encryption helper
	ctx, enc := loadEncryptionHelper(ctx, logger, cfg.Secret)

	// get core scheduler and inject into context
	ctx, coreScheduler := loadCoreScheduler(ctx)
	err = coreScheduler.Start()
	if err != nil {
		logger.Error("error on starting core scheduler", zap.Error(err))
		return err
	}

	// load bus plugin
	ctx, bus, err := loadBus(ctx, cfg.Bus)
	if err != nil {
		logger.Error("error on getting bus", zap.Error(err))
		return err
	}

	// load storage plugin
	ctx, storage, err := loadStorageDatabase(ctx, cfg.Database.Storage)
	if err != nil {
		logger.Error("error on getting storage database", zap.Error(err))
		return err
	}

	// load storage plugin
	ctx, metric, err := loadMetricDatabase(ctx, cfg.Database.Metric)
	if err != nil {
		logger.Error("error on getting metric database", zap.Error(err))
		return err
	}

	// load core api
	ctx, api, err := loadCoreApi(ctx)
	if err != nil {
		logger.Error("error on getting core api", zap.Error(err))
		return err
	}

	// add into struct
	s.ctx = ctx
	s.config = cfg
	s.logger = logger
	s.enc = enc
	s.coreSchedulerSVC = coreScheduler
	s.bus = bus
	s.storage = storage
	s.metric = metric
	s.api = api

	// perform restore operation
	s.checkSystemRestore()

	// TODO: perform internal upgrade

	// imports storage data from disk to database
	err = s.runStorageImport()
	if err != nil {
		return err
	}

	// setup initial system settings
	s.updateInitialSystemSettings()
	s.setupInitialUser()

	// start upgrade, if any
	s.startUpgrade()

	// engines needed in task and schedules service
	// get variables engine
	variablesEngine, err := s.getVariablesEngine()
	if err != nil {
		return err
	}

	// load services
	//-------------------------------------------

	// resource handler service
	resource, err := resourceSVC.New(ctx)
	if err != nil {
		logger.Error("error on getting resource handler service", zap.Error(err))
		return err
	}

	// message processor service
	messageProcessor, err := gwMsgProcessorSVC.New(ctx, "")
	if err != nil {
		logger.Error("error on getting message processor service", zap.Error(err))
		return err
	}

	// gateway service
	gateway, err := gatewaySVC.New(ctx, &cfg.Gateway)
	if err != nil {
		logger.Error("error on getting gateway service", zap.Error(err))
		return err
	}

	// task service
	task, err := taskSVC.New(ctx, &cfg.Task, variablesEngine)
	if err != nil {
		logger.Error("error on getting task service", zap.Error(err))
		return err
	}

	// scheduler service
	scheduler, err := schedulerSVC.New(ctx, &cfg.Schedule, variablesEngine, api.Sunrise())
	if err != nil {
		logger.Error("error on getting scheduler service", zap.Error(err))
		return err
	}

	// handler service
	handler, err := handlerSVC.New(ctx, &cfg.Handler)
	if err != nil {
		logger.Error("error on getting handler service", zap.Error(err))
		return err
	}

	// deletion service
	deletion, err := deletionSVC.New(ctx)
	if err != nil {
		logger.Error("error on getting deletion service", zap.Error(err))
		return err
	}

	// system job service
	systemJobs, err := systemJobsSVC.New(ctx)
	if err != nil {
		logger.Error("error on getting system jobs service", zap.Error(err))
		return err
	}

	// forward payload service
	forwardPayload, err := fwdPayloadSVC.New(ctx)
	if err != nil {
		logger.Error("error on getting forward payload service", zap.Error(err))
		return err
	}

	// websocket service
	websocket, err := websocketSVC.New(ctx, router)
	if err != nil {
		logger.Error("error on getting websocket service", zap.Error(err))
		return err
	}

	// virtual assistant service
	virtualAssistant, err := vaSVC.New(ctx, &cfg.VirtualAssistant, router)
	if err != nil {
		logger.Error("error on getting virtual assistant service", zap.Error(err))
		return err
	}

	// load http handlers
	// all other handlers will be registered prior to this call,
	// because this router registers a path with prefix "/" to handle static web assets
	// TODO: this to be fixed by adding only required paths for web contents
	httpHandler, err := httpRouter.New(ctx, cfg, router)
	if err != nil {
		logger.Error("error on getting http router", zap.Error(err))
		return err
	}

	// used for debug purpose, to see the registered paths
	s.printsHandlersPath(router)

	// http listener service
	httpListener, err := httpListenerSVC.New(ctx, cfg.Web, httpHandler)
	if err != nil {
		logger.Error("error on getting http listener service", zap.Error(err))
		return err
	}

	// load all the services in an array and start in order
	services := []serviceTY.Service{
		resource,
		messageProcessor,
		gateway,
		task,
		scheduler,
		handler,
		deletion,
		systemJobs,
		websocket,
		virtualAssistant,
		forwardPayload,
		// do not include http listener
	}

	// start all the services
	for index := range services {
		service := services[index]
		err := service.Start()
		if err != nil {
			s.logger.Error("error on starting a service", zap.String("name", service.Name()), zap.Error(err))
			return err
		}
	}

	// add services into struct
	s.resourceSVC = resource
	s.messageProcessorSVC = messageProcessor
	s.gatewaySVC = gateway
	s.taskSVC = task
	s.schedulerSVC = scheduler
	s.handlerSVC = handler
	s.deletionSVC = deletion
	s.systemJobsSVC = systemJobs
	s.websocketSVC = websocket
	s.virtualAssistantSVC = virtualAssistant
	s.fwdPayloadSVC = forwardPayload

	// call shutdown hook
	shutdownHook := NewShutdownHook(s.logger, s.stop, s.bus, true)
	go shutdownHook.Start()

	// start http handler service
	return httpListener.Start()
}

func (s *Server) stop() {
	// stop services, order of the execution is important
	services := []serviceTY.Service{
		s.fwdPayloadSVC,
		s.virtualAssistantSVC,
		s.websocketSVC,
		s.systemJobsSVC,
		s.deletionSVC,
		s.handlerSVC,
		s.schedulerSVC,
		s.taskSVC,
		s.gatewaySVC,
		s.messageProcessorSVC,
		s.resourceSVC,
		s.coreSchedulerSVC,
	}

	for index := range services {
		service := services[index]
		err := service.Close()
		if err != nil {
			s.logger.Error("error on closing a service", zap.String("name", service.Name()), zap.Error(err))
		}
	}

	// unload support apis and plugins
	if err := s.storage.Close(); err != nil {
		s.logger.Error("error on closing storage connection", zap.Error(err))
	}

	if err := s.metric.Close(); err != nil {
		s.logger.Error("error on closing metric connection", zap.Error(err))
	}

	if err := s.bus.Close(); err != nil {
		s.logger.Error("error on closing bus service", zap.Error(err))
	}
}

func (s *Server) getVariablesEngine() (types.VariablesEngine, error) {
	// engines needed in task and schedules service
	// create template engine with sunrise and sunset
	templateFuncsMap := map[string]interface{}{
		"sunrise": s.api.Sunrise().SunriseTime,
		"sunset":  s.api.Sunrise().SunsetTime,
		"version": version.Get,
	}
	// get template engine
	templateEngine, err := templateUtils.New(s.ctx, templateFuncsMap)
	if err != nil {
		s.logger.Error("error on getting template engine", zap.Error(err))
		return nil, err
	}

	// get variables engine
	variablesEngine, err := variablesUtils.New(s.ctx, templateEngine)
	if err != nil {
		s.logger.Error("error on getting variables engine", zap.Error(err))
		return nil, err
	}

	return variablesEngine, nil
}

func (s *Server) printsHandlersPath(router *mux.Router) {
	// execute only on debug enabled
	if !s.logger.Level().Enabled(zapcore.DebugLevel) {
		return
	}

	walkFunc := func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathRegexp()
		if err != nil {
			return err
		}

		s.logger.Debug("http handler registered path", zap.String("path", path))
		return nil
	}

	err := router.Walk(walkFunc)
	if err != nil {
		s.logger.Error("error on calling walk func", zap.Error(err))
	}

}

// starts the upgrade process
// checks the database for the version and applies the upgrade patches, if needed
func (s *Server) startUpgrade() {
	// get the version details from database
	sysVersion, err := s.api.Settings().GetVersion()
	if err != nil {
		s.logger.Fatal("error on getting version details", zap.Error(err))
	}
	s.logger.Info("checking for upgrade", zap.String("lastUpgradePatch", sysVersion.LastUpgrade))

	// if there is no version available, indicates that the software installed before the upgrade feature introduced
	// keep 1.9.9 as last update to proceed all the upgrades
	// Note: MyController V2 version starts from 2.0.0
	if sysVersion.LastUpgrade == "" {
		sysVersion.LastUpgrade = "1.9.9"
	}

	// trigger upgrade
	start := time.Now()
	appliedUpgradeVersion, err := upgrade.StartUpgrade(s.ctx, sysVersion.LastUpgrade)
	if err != nil {
		s.logger.Fatal("error on upgrade", zap.Error(err))
	}

	if appliedUpgradeVersion != "" {
		s.logger.Info("upgrade completed", zap.String("appliedPatchVersion", appliedUpgradeVersion), zap.String("timeTaken", time.Since(start).String()))
	}

	// update the version in to storage database
	sysVersion, err = s.api.Settings().GetVersion()
	if err != nil {
		s.logger.Fatal("error on getting version details", zap.Error(err))
	}
	ver := version.Get()
	sysVersion.Version = ver.Version
	sysVersion.GitCommit = ver.GitCommit
	sysVersion.Database = s.storage.Name()
	err = s.api.Settings().UpdateVersion(sysVersion)
	if err != nil {
		s.logger.Fatal("error on updating version details into database", zap.Error(err))
	}
}
