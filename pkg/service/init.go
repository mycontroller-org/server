package service

import (
	"errors"
	"flag"
	"io/ioutil"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	cfgml "github.com/mycontroller-org/backend/v2/pkg/model/config"
	"github.com/mycontroller-org/backend/v2/pkg/scheduler"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/version"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// services
var (
	CFG *cfgml.Config
	STG stgml.Client
	MTS mtsml.Client
	SCH *scheduler.Scheduler
)

// Init all the supported registries
func Init(preInitFunc func(), postInitFunc func(cfg *cfgml.Config)) {
	initConfig()
	initLogger()

	// trigger pre init func
	if preInitFunc != nil {
		preInitFunc()
	}

	initScheduler()
	initStorage()

	// trigger post init func
	if postInitFunc != nil {
		postInitFunc(CFG)
	}
}

// Close all the registries
func Close() error {
	// close scheduler
	if SCH != nil {
		SCH.Close()
	}

	// Close storage and metric database
	if STG != nil {
		err := STG.Close()
		if err != nil {
			zap.L().Error("Failed to close storage database")
		}
	}
	if MTS != nil {
		if MTS != nil {
			err := MTS.Close()
			if err != nil {
				zap.L().Error("Failed to close metrics database")
			}
		}
	}
	return nil
}

func initLogger() {
	logger := ut.GetLogger(CFG.Logger.Mode, CFG.Logger.Level.Core, CFG.Logger.Encoding, false, 0)
	zap.ReplaceGlobals(logger)
	zap.L().Info("Welcome to MyController.org server :)")
	zap.L().Info("Server detail", zap.Any("version", version.Get()), zap.Any("loggerConfig", CFG.Logger))
}

func initConfig() {
	// init a temp logger
	logger := ut.GetLogger("development", "error", "console", false, 0)

	cf := flag.String("config", "./config.yaml", "Configuration file")
	flag.Parse()
	if cf == nil {
		logger.Fatal("Configuration file not supplied")
	}
	d, err := ioutil.ReadFile(*cf)
	if err != nil {
		logger.Fatal("Error on reading configuration file", zap.Error(err))
	}

	err = yaml.Unmarshal(d, &CFG)
	if err != nil {
		logger.Fatal("Failed to unmarshal yaml data", zap.Error(err))
	}
}

func initScheduler() {
	SCH = scheduler.Init()
	SCH.Start()
}

func initStorage() {
	// Get storage and metric database config
	storageCfg, err := getDatabaseConfig(CFG.Database.Storage)
	if err != nil {
		zap.L().Fatal("Problem with storage database config", zap.String("name", CFG.Database.Storage), zap.Error(err))
	}

	metricsCfg, err := getDatabaseConfig(CFG.Database.Metrics)
	if err != nil {
		zap.L().Fatal("Problem with metrics database config", zap.String("name", CFG.Database.Metrics), zap.Error(err))
	}

	// include logger details
	storageCfg["logger"] = map[string]string{"mode": CFG.Logger.Mode, "encoding": CFG.Logger.Encoding, "level": CFG.Logger.Level.Storage}
	metricsCfg["logger"] = map[string]string{"mode": CFG.Logger.Mode, "encoding": CFG.Logger.Encoding, "level": CFG.Logger.Level.Metrics}

	// Init storage database
	STG, err = InitStorageDatabase(storageCfg, SCH)
	if err != nil {
		zap.L().Fatal("Error on storage db init", zap.Error(err))
	}

	// Init metrics database
	MTS, err = InitMetricsDatabase(metricsCfg)
	if err != nil {
		zap.L().Fatal("Error on metrics db init", zap.Error(err))
	}
}

func getDatabaseConfig(name string) (map[string]interface{}, error) {
	for _, d := range CFG.Databases {
		if d["name"] == name {
			return d, nil
		}
	}
	return nil, errors.New("Config not found")
}

// GetDirDataRoot returns data root location
func GetDirDataRoot() string {
	if CFG.Directories.Data == "" {
		return model.DefaultDirDataRoot
	}
	return CFG.Directories.Data
}

// GetDirLogsRoot returns logs root location
func GetDirLogsRoot() string {
	if CFG.Directories.Logs == "" {
		return model.DefaultDirLogsRoot
	}
	return CFG.Directories.Logs
}
