package service

import (
	"errors"
	"flag"
	"io/ioutil"

	"github.com/mustafaturan/bus"
	"github.com/mustafaturan/monoton"
	"github.com/mustafaturan/monoton/sequencer"
	"github.com/mycontroller-org/mycontroller-v2/pkg/storage"
	ms "github.com/mycontroller-org/mycontroller-v2/pkg/storage/metrics"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// services
var (
	CFG *Config
	BUS *bus.Bus
	STG storage.Client
	MTS ms.Client
)

// Init all the supported registries
func Init() {
	initLogger()
	initConfig()
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
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	zap.L().Info("Welcome to MyController 2.x :)")
}

func initConfig() {
	cf := flag.String("config", "./config.yaml", "Configuration file")
	flag.Parse()
	if cf == nil {
		zap.L().Fatal("Configuration file not supplied")
	}
	zap.L().Debug("Configuration file path:", zap.String("file", *cf))
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
	s, err := storage.Init(sCFG)
	if err != nil {
		zap.L().Fatal("Error on storage db init", zap.Error(err))
	}
	STG = *s

	// Init metrics database
	ms, err := ms.Init(mCFG)
	if err != nil {
		zap.L().Fatal("Error on metrics db init", zap.Error(err))
	}
	MTS = *ms
}

func getDatabaseConfig(name string) (map[string]interface{}, error) {
	for _, d := range CFG.Databases {
		if d["name"] == name {
			return d, nil
		}
	}
	return nil, errors.New("Config not found")
}
