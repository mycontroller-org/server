package configuration

import (
	"flag"
	"io/ioutil"

	cfgml "github.com/mycontroller-org/backend/v2/pkg/model/config"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// configuration globally accessable
var (
	CFG *cfgml.Config
)

// Init configuration
func Init() {
	// init a temp logger
	logger := utils.GetLogger("development", "error", "console", false, 0)

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
