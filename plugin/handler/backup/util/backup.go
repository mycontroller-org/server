package backup

import (
	"context"
	"fmt"
	"time"

	bkpMap "github.com/mycontroller-org/server/v2/pkg/backup"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/ziputils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	backupAPI "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Backup creates zip file on a tmp location and returns the location details
func Backup(ctx context.Context, logger *zap.Logger, prefix, storageExportType string, includeSecureShare, includeInsecureShare bool, storage storageTY.Plugin, bus busTY.Plugin) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	dstDir := fmt.Sprintf("%s/%s_%s_%s_%s", types.GetEnvString(types.ENV_DIR_DATA_STORAGE), prefix, BackupIdentifier, storageExportType, timestamp)
	zipFilename := fmt.Sprintf("%s.zip", dstDir)

	// get backup directories
	bkpDirectories, err := bkpMap.GetDirectories(includeSecureShare, includeInsecureShare)
	if err != nil {
		logger.Error("error on get backup directories", zap.Error(err))
		return "", err
	}

	// create new backup instance
	backupRestore, err := backupAPI.New(ctx, bkpDirectories)
	if err != nil {
		return "", err
	}
	exportFuncMap, err := bkpMap.GetStorageApiMap(ctx)
	if err != nil {
		return "", err
	}

	// exports storage and directories
	err = backupRestore.ExportStorage(exportFuncMap, bkpMap.MyControllerDataTransformationExportFunc, dstDir, storageExportType)
	if err != nil {
		return "", err
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
