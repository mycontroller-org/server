package importexport

import (
	"fmt"
	"strings"

	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	backupAPI "github.com/mycontroller-org/server/v2/pkg/backup"
	"github.com/mycontroller-org/server/v2/pkg/json"
	backupML "github.com/mycontroller-org/server/v2/pkg/model/backup"
	handlerML "github.com/mycontroller-org/server/v2/pkg/model/handler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	"github.com/mycontroller-org/server/v2/plugin/handler/backup/disk"
	backupUtil "github.com/mycontroller-org/server/v2/plugin/handler/backup/util"
	"go.uber.org/zap"
)

// RunRestore func
func RunRestore(file backupML.BackupFile) error {
	zap.L().Info("restore data request received", zap.Any("backupFile", file))

	err := backupAPI.ExtractExportedZipfile(file.FullPath)
	if err != nil {
		zap.L().Error("error on extract", zap.Error(err), zap.Any("file", file))
		return err
	}
	// TODO: do shutdown
	zap.L().Info("all set to start restore on startup. the server is going to down now. start the server manually")
	busUtils.PostShutdownEvent()

	return nil
}

// RunOnDemandBackup triggers on demand export
func RunOnDemandBackup(input *backupML.OnDemandBackupConfig) error {
	zap.L().Debug("on-demand backup request received", zap.Any("config", input))

	configData := disk.Config{
		Prefix:            fmt.Sprintf("%s_on_demand", input.Prefix),
		StorageExportType: input.StorageExportType,
		TargetDirectory:   input.TargetLocation,
		RetentionCount:    0,
	}

	exporterData := handlerML.BackupData{
		ProviderType: backupUtil.ProviderDisk,
		Spec:         utils.StructToMap(configData),
	}

	base64String, err := yamlUtils.MarshalBase64Yaml(&exporterData)
	if err != nil {
		zap.L().Error("error on converting exporter data to base64", zap.Error(err))
		return err
	}

	data := handlerML.GenericData{
		Disabled: "false",
		Type:     handlerML.DataTypeBackup,
		Data:     base64String,
	}

	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	finalData := map[string]string{
		"on_demand_backup": string(dataBytes),
	}
	busUtils.PostToHandler([]string{input.Handler}, finalData)
	return nil
}

// GetBackupFilesList details
func GetBackupFilesList() ([]interface{}, error) {
	locationsSettings, err := settingsAPI.GetBackupLocations()
	if err != nil {
		return nil, err
	}

	locations := locationsSettings.Locations

	exportedFiles := make([]interface{}, 0)

	for _, location := range locations {
		if location.Type == backupUtil.ProviderDisk {
			diskLocation := &backupML.BackupLocationDisk{}
			err = utils.MapToStruct(utils.TagNameNone, location.Config, diskLocation)
			if err != nil {
				return exportedFiles, err
			}
			rawFiles, err := utils.ListFiles(diskLocation.TargetDirectory)
			if err != nil {
				return exportedFiles, err
			}
			for _, rawFile := range rawFiles {
				if rawFile.IsDir || !strings.Contains(rawFile.Name, backupUtil.BackupIdentifier) {
					continue
				}
				exportedFile := backupML.BackupFile{
					ID:           rawFile.FullPath,
					LocationName: location.Name,
					ProviderType: location.Type,
					Directory:    diskLocation.TargetDirectory,
					FileName:     rawFile.Name,
					FileSize:     rawFile.Size,
					FullPath:     rawFile.FullPath,
					ModifiedOn:   rawFile.ModifiedTime,
				}
				exportedFiles = append(exportedFiles, exportedFile)
			}
		}
	}

	return exportedFiles, nil
}
