package config

import (
	"path"
)

// Files, directory locations
const (
	DefaultDirDataRoot = "/tmp/mc/data" // default data dir location
	DefaultDirLogsRoot = "/tmp/mc/logs" // default logs dir location
	DefaultDirTmp      = "/tmp/mc/tmp"  // default tmp dir location

	DirectoryDataFirmware = "/firmware"     // location to keep firmware files
	DirectoryDataStorage  = "/storage"      // location to keep storage database exported files
	DirectoryDataInternal = "/internal"     // location to keep system internal files
	DirectoryGatewayLogs  = "/gateway_logs" // location to keep gateway message logs
	DirectoryGatewayTmp   = "/gateway_tmp"  // location to keep gateway related tmp items
)

// return data root directory location
func (dr *Directories) GetData() string {
	if dr.Data != "" {
		return dr.Data
	}
	return DefaultDirDataRoot
}

// return logs root directory location
func (dr *Directories) GetLogs() string {
	if dr.Logs != "" {
		return dr.Logs
	}
	return DefaultDirLogsRoot
}

// returns temporary directory location
func (dr *Directories) GetTmp() string {
	if dr.Tmp != "" {
		return dr.Tmp
	}
	return DefaultDirTmp
}

// GetDataFirmware location
func (dr *Directories) GetDataFirmware() string {
	return getDirectoryFullPath(dr.GetData(), DirectoryDataFirmware)
}

// GetGatewayLogs location
func (dr *Directories) GetGatewayLogs() string {
	return getDirectoryFullPath(dr.GetLogs(), DirectoryGatewayLogs)
}

// GetDataStorage location
func (dr *Directories) GetDataStorage() string {
	return getDirectoryFullPath(dr.GetData(), DirectoryDataStorage)
}

// GetDirectoryStorage location
func (dr *Directories) GetDataInternal() string {
	return getDirectoryFullPath(dr.GetData(), DirectoryDataInternal)
}

func (dr *Directories) GetGatewayTmp() string {
	return getDirectoryFullPath(dr.GetData(), DirectoryGatewayTmp)
}

// helper
func getDirectoryFullPath(paths ...string) string {
	return path.Join(paths...)
}
