package model

import (
	"path"

	"github.com/mycontroller-org/server/v2/pkg/model/config"
)

// Files, directory locations
const (
	DefaultDirDataRoot = "/tmp/mc/data" // default data dir location
	DefaultDirLogsRoot = "/tmp/mc/logs" // default logs dir location
	DefaultDirTmp      = "/tmp/mc/tmp"  // default tmp dir location

	DirectoryDataFirmware       = "/firmware"         // location to keep firmware files
	DirectoryDataStorage        = "/storage"          // location to keep storage database exported files
	DirectoryDataInternal       = "/internal"         // location to keep system internal files
	DirectoryLogsGateway        = "/gateway_logs"     // location to keep gateway message logs
	DirectoryTmpGatewayFirmware = "/gateway/firmware" // location to keep gateway related tmp items
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

func getDirectoryFullPath(paths ...string) string {
	return path.Join(paths...)
}

// GetDataDirectoryFirmware location
func GetDataDirectoryFirmware() string {
	return getDirectoryFullPath(dir.Data, DirectoryDataFirmware)
}

// GetLogsDirectoryGatewayLog location
func GetLogsDirectoryGatewayLog() string {
	return getDirectoryFullPath(dir.Logs, DirectoryLogsGateway)
}

// GetDataDirectoryStorage location
func GetDataDirectoryStorage() string {
	return getDirectoryFullPath(dir.Data, DirectoryDataStorage)
}

// GetDirectoryStorage location
func GetDataDirectoryInternal() string {
	return getDirectoryFullPath(dir.Data, DirectoryDataInternal)
}

func GetTmpGatewayFirmware() string {
	return getDirectoryFullPath(dir.Tmp, DirectoryTmpGatewayFirmware)
}
