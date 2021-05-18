package configuration

import (
	"flag"
	"fmt"
	"io/ioutil"

	cfgml "github.com/mycontroller-org/backend/v2/pkg/model/config"
	loggerUtils "github.com/mycontroller-org/backend/v2/pkg/utils/logger"
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
	logger := loggerUtils.GetLogger("development", "error", "console", false, 0)

	cf := flag.String("config", "./config.yaml", "Configuration file")
	flag.Parse()
	if cf == nil {
		logger.Fatal("Configuration file not supplied")
		return
	}
	d, err := ioutil.ReadFile(*cf)
	if err != nil {
		logger.Fatal("Error on reading configuration file", zap.Error(err))
	}

	err = yaml.Unmarshal(d, &CFG)
	if err != nil {
		logger.Fatal("Failed to unmarshal yaml data", zap.Error(err))
	}

	// update encryption key length
	// converts it to fixed size as 32 bytes
	CFG.Secret = updatedKey(CFG.Secret)
}

// UpdatedKey returns fixed key size
// that is 32 bytes
func updatedKey(actualKey string) string {
	return fmt.Sprintf("%032.32s", actualKey)
}
