package configuration

import (
	"errors"
	"fmt"
	"os"

	cfgTY "github.com/mycontroller-org/server/v2/pkg/types/config"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// returns configuration
func Get(configFilename string) (*cfgTY.Config, error) {
	// load a temporary logger
	logger := loggerUtils.GetLogger(loggerUtils.ModeRecordAll, "error", "console", false, 0, true)

	if configFilename == "" {
		logger.Fatal("configuration file not supplied")
		return nil, errors.New("configuration file not supplied")
	}
	d, err := os.ReadFile(configFilename)
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
	CFG.Secret = updatedKey(logger, CFG.Secret)

	return &CFG, nil
}

// UpdatedKey returns fixed key size
// that is 32 bytes
func updatedKey(logger *zap.Logger, actualKey string) string {
	if len(actualKey) > 32 {
		logger.Warn("secret length is greater than 32 characters. takes only the first 32 characters", zap.Int("length", len(actualKey)))
	}
	return fmt.Sprintf("%032.32s", actualKey)
}
