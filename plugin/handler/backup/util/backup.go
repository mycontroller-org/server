package backup

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	backupAPI "github.com/mycontroller-org/server/v2/pkg/backup"
	bkpMap "github.com/mycontroller-org/server/v2/pkg/backup/bkp_map"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/ziputils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Backup creates zip file on a tmp location and returns the location details
func Backup(ctx context.Context, logger *zap.Logger, prefix, storageExportType string, includeSecureShare, includeInsecureShare bool, storage storageTY.Plugin, bus busTY.Plugin) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	dstDir := fmt.Sprintf("%s/%s_%s_%s_%s", types.GetEnvString(types.ENV_DIR_DATA_STORAGE), prefix, BackupIdentifier, storageExportType, timestamp)
	zipFilename := fmt.Sprintf("%s.zip", dstDir)

	// create new backup instance
	backupRestore, err := backupAPI.New(ctx)
	if err != nil {
		return "", err
	}
	exportFuncMap, err := bkpMap.GetStorageApiMap(ctx)
	if err != nil {
		return "", err
	}
	err = backupRestore.ExportStorage(exportFuncMap, dstDir, storageExportType)
	if err != nil {
		return "", err
	}

	// copy firmware files
	baseDir := types.GetEnvString(types.ENV_DIR_DATA_FIRMWARE)
	if baseDir == "" {
		return "", fmt.Errorf("environment '%s' not set", types.ENV_DIR_DATA_FIRMWARE)
	}
	err = backupRestore.CopyFiles(baseDir, dstDir, false)
	if err != nil {
		return "", err
	}

	// copy shared directory: secure and insecure
	if includeSecureShare {
		secureShareDir := types.GetEnvString(types.ENV_DIR_SHARE_SECURE)
		if secureShareDir == "" {
			return "", fmt.Errorf("environment '%s' not set", types.ENV_DIR_SHARE_SECURE)
		}
		dstSecureShareDir := filepath.Join(dstDir, config.DirectorySecureShare)
		err = backupRestore.CopyFiles(secureShareDir, dstSecureShareDir, false)
		if err != nil {
			return "", err
		}
	}

	// copy shared directory: insecure
	if includeInsecureShare {
		inSecureShareDir := types.GetEnvString(types.ENV_DIR_SHARE_INSECURE)
		if inSecureShareDir == "" {
			return "", fmt.Errorf("environment '%s' not set", types.ENV_DIR_SHARE_INSECURE)
		}
		dstInsecureShareDir := filepath.Join(dstDir, config.DirectoryInsecureShare)
		err = backupRestore.CopyFiles(inSecureShareDir, dstInsecureShareDir, false)
		if err != nil {
			return "", err
		}
	}

	// create zip from exported directory
	err = ziputils.Zip(logger, dstDir, zipFilename)
	if err != nil {
		return "", err
	}

	// remove tmp location
	err = utils.RemoveDir(dstDir)
	if err != nil {
		logger.Error("error on removing backup tmp location", zap.Error(err), zap.String("backupTmpLocation", dstDir))
	}

	return zipFilename, nil
}
