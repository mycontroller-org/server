package types

import (
	"fmt"
	"os"
)

const (
	ENV_TELEMETRY_ENABLED      = "MC_TELEMETRY_ENABLED"
	ENV_DOCUMENTATION_URL      = "MC_DOCUMENTATION_URL"
	ENV_METRIC_DB_DISABLED     = "MC_METRIC_DB_DISABLED"
	ENV_DIR_DATA               = "MC_DIR_DATA"
	ENV_DIR_DATA_STORAGE       = "MC_DIR_DATA_STORAGE"
	ENV_DIR_DATA_FIRMWARE      = "MC_DIR_DATA_FIRMWARE"
	ENV_DIR_DATA_INTERNAL      = "MC_DIR_DATA_INTERNAL"
	ENV_DIR_TMP                = "MC_DIR_TMP"
	ENV_DIR_GATEWAY_LOGS       = "MC_DIR_GATEWAY_LOGS"
	ENV_DIR_GATEWAY_TMP        = "MC_DIR_GATEWAY_TMP"
	ENV_LOG_LEVEL_CORE         = "MC_LOG_LEVEL_CORE"
	ENV_LOG_LEVEL_STORAGE      = "MC_LOG_LEVEL_STORAGE"
	ENV_LOG_LEVEL_METRIC       = "MC_LOG_LEVEL_METRIC"
	ENV_LOG_LEVEL_WEB_HANDLER  = "MC_LOG_LEVEL_WEB_HANDLER"
	ENV_LOG_ENABLE_STACK_TRACE = "MC_LOG_ENABLE_STACK_TRACE"
	ENV_LOG_ENCODING           = "MC_LOG_ENCODING"
	ENV_LOG_MODE               = "MC_LOG_MODE"

	ENV_JWT_ACCESS_SECRET = "JWT_ACCESS_SECRET" // environment variable to set secret for JWT token
)

func GetEnvBool(key string) bool {
	return os.Getenv(key) == "true"
}

func GetEnvString(key string) string {
	return os.Getenv(key)
}

func SetEnv(key string, value interface{}) error {
	return os.Setenv(key, fmt.Sprintf("%v", value))
}
