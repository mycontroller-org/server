package exporter

import (
	"fmt"
	"time"

	exportAPI "github.com/mycontroller-org/backend/v2/pkg/export"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/ziputils"
	"go.uber.org/zap"
)

// Export creates zip file on a tmp location and returns the location details
func Export(prefix, exportType string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	targetDir := fmt.Sprintf("%s/%s_%s_%s_%s", model.GetDirectoryTmp(), prefix, ExportIdentifier, exportType, timestamp)
	zipFilename := fmt.Sprintf("%s.zip", targetDir)

	// export to tmp directory
	err := exportAPI.ExecuteExport(targetDir, exportType)
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
		zap.L().Error("error on removing export tmp location", zap.Error(err), zap.String("exportTmpLocation", targetDir))
	}

	return zipFilename, nil
}
