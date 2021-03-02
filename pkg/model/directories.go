package model

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model/config"
)

// Files, directory locations
const (
	DefaultDirDataRoot = "/tmp/mc/data" // default data dir location
	DefaultDirLogsRoot = "/tmp/mc/logs" // default data dir location
	DefaultDirTmp      = "/tmp/mc/tmp"

	DirectoryFirmwares          = "/firmwares"            // location to keep firmware files
	DirectoryGatewayMessageLogs = "/gateway_message_logs" // location to keep gateway message logs
	DirectoryExports            = "/exports"              // location to keep storage database exported files
)

// dir reference should be loaded at startup
var dir config.Directories

// UpdateDirecotries updates dir locations
func UpdateDirecotries(cfg config.Directories) {
	dir = config.Directories{
		Data: DefaultDirDataRoot,
		Logs: DefaultDirLogsRoot,
		Tmp:  DefaultDirTmp,
	}
	if cfg.Data != "" {
		dir.Data = cfg.Data
	}
	if cfg.Logs != "" {
		dir.Logs = cfg.Logs
	}

	if cfg.Tmp != "" {
		dir.Tmp = cfg.Tmp
	}
}

// GetDirectoryDataRoot location
func GetDirectoryDataRoot() string {
	return dir.Data
}

// GetDirectoryLogsRoot location
func GetDirectoryLogsRoot() string {
	return dir.Logs
}

// GetDirectoryTmp location
func GetDirectoryTmp() string {
	return dir.Tmp
}

// GetDirectoryFirmware location
func GetDirectoryFirmware() string {
	return getDirectoryFullPath(dir.Data, DirectoryFirmwares)
}

// GetDirectoryGatewayLog location
func GetDirectoryGatewayLog() string {
	return getDirectoryFullPath(dir.Logs, DirectoryGatewayMessageLogs)
}

// GetDirectoryExport location
func GetDirectoryExport() string {
	return getDirectoryFullPath(dir.Data, DirectoryExports)
}

func getDirectoryFullPath(rootDir, subDir string) string {
	return fmt.Sprintf("%s%s", rootDir, subDir)
}
