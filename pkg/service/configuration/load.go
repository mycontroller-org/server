package configuration

import (
	"errors"
	"flag"
	"fmt"
	"os"

	cfgTY "github.com/mycontroller-org/server/v2/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// configuration globally accessable
var (
	PauseModifiedOnUpdate = concurrency.SafeBool{}
)

// Load configuration
func Load() (*cfgTY.Config, error) {
	// load a temporary logger
	logger := loggerUtils.GetLogger("development", "error", "console", false, 0, true)

	cf := flag.String("config", "./config.yaml", "Configuration file")
	flag.Parse()
	if cf == nil {
		logger.Fatal("configuration file not supplied")
		return nil, errors.New("configuration file not supplied")
	}
	d, err := os.ReadFile(*cf)
	if err != nil {
		logger.Fatal("error on reading configuration file", zap.Error(err))
	}

	CFG := cfgTY.Config{}
	err = yaml.Unmarshal(d, &CFG)
	if err != nil {
		logger.Fatal("failed to parse yaml data", zap.Error(err))
		return nil, err
	}

	// verify secret availability
	if CFG.Secret == "" {
		logger.Fatal("empty secret is not allowed")
	}

	// update encryption key length
	// converts it to fixed size as 32 bytes
	CFG.Secret = updatedKey(CFG.Secret)

	// load default value
	PauseModifiedOnUpdate.Reset()
	return &CFG, nil
}

// UpdatedKey returns fixed key size
// that is 32 bytes
func updatedKey(actualKey string) string {
	if len(actualKey) > 32 {
		zap.L().Warn("secret length is greater than 32 characters. takes only the first 32 characters", zap.Int("length", len(actualKey)))
	}
	return fmt.Sprintf("%032.32s", actualKey)
}
