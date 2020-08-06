package service

import (
	"errors"
	"flag"
	"io/ioutil"

	"github.com/mustafaturan/bus"
	"github.com/mustafaturan/monoton"
	"github.com/mustafaturan/monoton/sequencer"
	"github.com/mycontroller-org/mycontroller-v2/pkg/model/config"
	"github.com/mycontroller-org/mycontroller-v2/pkg/storage"
	ms "github.com/mycontroller-org/mycontroller-v2/pkg/storage/metrics"
	"github.com/mycontroller-org/mycontroller-v2/pkg/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// services
var (
	CFG *config.Config
	BUS *bus.Bus
	STG storage.Client
	MTS ms.Client
)

// Init all the supported registries
func Init() {
	initConfig()
	initLogger()
	initBus()
	initStorage()
}

// Close all the registries
func Close() error {
	// deregister handlers
	for _, hk := range BUS.HandlerKeys() {
		BUS.DeregisterHandler(hk)
	}
	// deregister topics
	for _, t := range BUS.Topics() {
		BUS.DeregisterTopics(t)
	}
	return nil
}

func initLogger() {
	logger := util.GetLogger(CFG.Logger.Level.Core, CFG.Logger.Encoding, false, 0)
	zap.ReplaceGlobals(logger)
	zap.L().Info("Welcome to MyController.org server :)")
	zap.L().Debug("Logger settings", zap.Any("loggerConfig", CFG.Logger))
}

func initConfig() {
	// init a temp logger
	logger := util.GetLogger("error", "console", false, 0)
	zap.ReplaceGlobals(logger)

	cf := flag.String("config", "./config.yaml", "Configuration file")
	flag.Parse()
	if cf == nil {
		zap.L().Fatal("Configuration file not supplied")
	}
	d, err := ioutil.ReadFile(*cf)
	if err != nil {
		zap.L().Fatal("Error on reading configuration file", zap.Error(err))
	}

	err = yaml.Unmarshal(d, &CFG)
	if err != nil {
		zap.L().Fatal("Failed to unmarshal yaml data", zap.Error(err))
	}
}

func initBus() {
	node := uint64(1)
	initialTime := uint64(1577865600000) // set 2020-01-01 PST as initial time
	m, err := monoton.New(sequencer.NewMillisecond(), node, initialTime)
	if err != nil {
		zap.L().Fatal("Error on creating bus", zap.Error(err))
	}
	// init an id generator
	var idGenerator bus.Next = (*m).Next
	// create a new bus instance
	b, err := bus.NewBus(idGenerator)
	if err != nil {
		zap.L().Fatal("Error on creating bus", zap.Error(err))
	}
	BUS = b
}

func initStorage() {
	// Get storage and metric database config
	sCFG, err := getDatabaseConfig(CFG.Database.Storage)
	if err != nil {
		zap.L().Fatal("Problem with storage database config", zap.String("name", CFG.Database.Storage), zap.Error(err))
	}
	mCFG, err := getDatabaseConfig(CFG.Database.Metrics)
	if err != nil {
		zap.L().Fatal("Problem with metrics database config", zap.String("name", CFG.Database.Metrics), zap.Error(err))
	}

	// Init storage database
	STG, err = storage.Init(sCFG)
	if err != nil {
		zap.L().Fatal("Error on storage db init", zap.Error(err))
	}

	// Init metrics database
	MTS, err = ms.Init(mCFG)
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
