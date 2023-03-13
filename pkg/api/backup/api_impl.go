package importexport

import (
	"fmt"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	"github.com/mycontroller-org/server/v2/plugin/handler/backup/disk"
	backupUtil "github.com/mycontroller-org/server/v2/plugin/handler/backup/util"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// RunRestore func
func (bk *BackupAPI) RunRestore(file backupTY.BackupFile) error {
	bk.logger.Info("restore data request received", zap.Any("backupFile", file))

	err := bk.backupRestore.ExtractExportedZipFile(file.FullPath)
	if err != nil {
		bk.logger.Error("error on extract", zap.Error(err), zap.Any("file", file))
		return err
	}
	// TODO: do shutdown
	bk.logger.Info("all set to start restore on startup. the server is going to down now. start the server manually")
	busUtils.PostShutdownEvent(bk.logger, bk.bus)

	return nil
}

// triggers on demand export
func (bk *BackupAPI) RunOnDemandBackup(input *backupTY.OnDemandBackupConfig) error {
	bk.logger.Debug("on-demand backup request received", zap.Any("config", input))

	configData := disk.Config{
		Prefix:            fmt.Sprintf("%s_on_demand", input.Prefix),
		StorageExportType: input.StorageExportType,
		TargetDirectory:   input.TargetLocation,
		RetentionCount:    0,
	}

	exporterData := handlerTY.BackupData{
		ProviderType: backupUtil.ProviderDisk,
		Spec:         utils.StructToMap(configData),
	}

	base64String, err := yamlUtils.MarshalBase64Yaml(&exporterData)
	if err != nil {
		bk.logger.Error("error on converting exporter data to base64", zap.Error(err))
		return err
	}

	data := handlerTY.GenericData{
		Disabled: "false",
		Type:     handlerTY.DataTypeBackup,
		Data:     base64String,
	}

	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	finalData := map[string]string{
		"on_demand_backup": string(dataBytes),
	}
	busUtils.PostToHandler(bk.logger, bk.bus, []string{input.Handler}, finalData)
	return nil
}

// GetBackupFilesList details
func (bk *BackupAPI) GetBackupFilesList() ([]interface{}, error) {
	locationsSettings, err := bk.settingsAPI.GetBackupLocations()
	if err != nil {
		return nil, err
	}

	locations := locationsSettings.Locations

	exportedFiles := make([]interface{}, 0)

	for _, location := range locations {
		if location.Type == backupUtil.ProviderDisk {
			diskLocation := &backupTY.BackupLocationDisk{}
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
				exportedFile := backupTY.BackupFile{
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
