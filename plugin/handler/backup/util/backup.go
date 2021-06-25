package backup

import (
	"fmt"
	"time"

	backupAPI "github.com/mycontroller-org/server/v2/pkg/backup"
	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/ziputils"
	"go.uber.org/zap"
)

// Backup creates zip file on a tmp location and returns the location details
func Backup(prefix, storageExportType string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	targetDir := fmt.Sprintf("%s/%s_%s_%s_%s", model.GetDirectoryTmp(), prefix, BackupIdentifier, storageExportType, timestamp)
	zipFilename := fmt.Sprintf("%s.zip", targetDir)

	// export to tmp directory
	err := backupAPI.ExecuteExportStorage(targetDir, storageExportType)
	if err != nil {
		return "", err
	}

	// copy firmware files
	err = backupAPI.ExecuteCopyFirmware(targetDir)
	if err != nil {
		return "", err
	}

	// create zip from exported directory
	err = ziputils.Zip(targetDir, zipFilename)
	if err != nil {
		return "", err
	}

	// remove tmp location
	err = utils.RemoveDir(targetDir)
	if err != nil {
		zap.L().Error("error on removing backup tmp location", zap.Error(err), zap.String("backupTmpLocation", targetDir))
	}

	return zipFilename, nil
}
