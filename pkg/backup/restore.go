package export

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	dashboardAPI "github.com/mycontroller-org/server/v2/pkg/api/dashboard"
	dataRepositoryAPI "github.com/mycontroller-org/server/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	fwAPI "github.com/mycontroller-org/server/v2/pkg/api/firmware"
	fwdPayloadAPI "github.com/mycontroller-org/server/v2/pkg/api/forward_payload"
	gwAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	notificationHandlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	sourceAPI "github.com/mycontroller-org/server/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	userAPI "github.com/mycontroller-org/server/v2/pkg/api/user"
	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	dataRepositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/server/v2/pkg/utils/ziputils"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var isImportJobRunning = concurrency.SafeBool{}

func ExecuteRestore(extractedDir string) error {
	start := time.Now()
	zap.L().Info("Restore job triggered", zap.String("extractedDirectory", extractedDir))

	err := store.STORAGE.Pause()
	if err != nil {
		zap.L().Fatal("error on pause a database", zap.Error(err))
		return err
	}

	err = store.STORAGE.ClearDatabase()
	if err != nil {
		zap.L().Fatal("error on emptying database", zap.Error(err))
		return err
	}

	storageDir := path.Join(extractedDir, types.DirectoryDataStorage)
	firmwareDir := path.Join(extractedDir, types.DirectoryDataFirmware)

	dataBytes, err := utils.ReadFile(extractedDir, backupTY.BackupDetailsFilename)
	if err != nil {
		zap.L().Fatal("error on reading export details", zap.String("dir", extractedDir), zap.String("filename", backupTY.BackupDetailsFilename), zap.Error(err))
		return err
	}

	exportDetails := &backupTY.BackupDetails{}
	err = yaml.Unmarshal(dataBytes, exportDetails)
	if err != nil {
		zap.L().Fatal("error on loading export details", zap.Error(err))
		return err
	}

	err = ExecuteImportStorage(storageDir, exportDetails.StorageExportType, false)
	if err != nil {
		zap.L().Fatal("error on importing storage files", zap.Error(err))
		return err
	}

	err = store.STORAGE.Resume()
	if err != nil {
		zap.L().Fatal("error on resume a database service", zap.Error(err))
		return err
	}

	zap.L().Info("Import database completed", zap.String("timeTaken", time.Since(start).String()))

	// restore firmwares
	err = ExecuteRestoreFirmware(firmwareDir)
	if err != nil {
		zap.L().Fatal("error on copying firmware files", zap.Error(err))
		return err
	}
	zap.L().Info("restore completed successfully", zap.String("timeTaken", time.Since(start).String()))
	return nil
}

// ExecuteRestoreFirmware copies firmwares to the actual directory
func ExecuteRestoreFirmware(sourceDir string) error {
	if isImportJobRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isImportJobRunning.Set()
	defer isImportJobRunning.Reset()

	destDir := types.GetDataDirectoryFirmware()
	err := utils.RemoveDir(destDir)
	if err != nil {
		return err
	}

	if !utils.IsDirExists(sourceDir) {
		return nil
	}

	err = os.Rename(sourceDir, destDir)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteImportStorage update data into database
func ExecuteImportStorage(sourceDir, fileType string, ignoreEmptyDir bool) error {
	if isImportJobRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isImportJobRunning.Set()
	defer func() {
		isImportJobRunning.Reset()
		// resume bus service
		mcbus.Resume()
	}()

	// pause bus service
	mcbus.Pause()

	zap.L().Info("Executing import job", zap.String("sourceDir", sourceDir), zap.String("fileType", fileType))
	// check directory availability
	if !utils.IsDirExists(sourceDir) {
		return fmt.Errorf("specified directory not available. sourceDir:%s", sourceDir)
	}
	files, err := utils.ListFiles(sourceDir)
	if err != nil {
		zap.L().Error("error on listing files", zap.Error(err))
		return err
	}
	if len(files) == 0 {
		if ignoreEmptyDir {
			return nil
		}
		return fmt.Errorf("no files found on:%s", sourceDir)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name, fileType) {
			continue
		}
		entityName := getEntityName(file.Name)
		zap.L().Debug("Importing a file", zap.String("file", file.Name), zap.String("entityName", entityName))
		// read data from file
		fileBytes, err := utils.ReadFile(sourceDir, file.Name)
		if err != nil {
			zap.L().Error("error on reading a file", zap.String("fileName", file.FullPath), zap.Error(err))
			return err
		}
		err = updateEntities(fileBytes, entityName, fileType)
		if err != nil {
			zap.L().Error("error on updating entity", zap.String("file", file.FullPath), zap.String("entityName", entityName), zap.String("fileType", fileType), zap.Error(err))
			return err
		}
	}
	return nil
}

func updateEntities(fileBytes []byte, entityName, fileFormat string) error {
	switch entityName {
	case types.EntityGateway:
		entities := make([]gatewayTY.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = gwAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntityNode:
		entities := make([]nodeTY.Node, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = nodeAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntitySource:
		entities := make([]sourceTY.Source, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = sourceAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntityField:
		entities := make([]fieldTY.Field, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = fieldAPI.Save(&entities[index], false)
			if err != nil {
				return err
			}
		}

	case types.EntityFirmware:
		entities := make([]firmwareTY.Firmware, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = fwAPI.Save(&entities[index], false)
			if err != nil {
				return err
			}
		}

	case types.EntityUser:
		entities := make([]userTY.User, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = userAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntityDashboard:
		entities := make([]dashboardTY.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = dashboardAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntityHandler:
		entities := make([]handlerTY.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = notificationHandlerAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntityForwardPayload:
		entities := make([]fwdPayloadTY.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = fwdPayloadAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntityTask:
		entities := make([]taskTY.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = taskAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntitySchedule:
		entities := make([]scheduleTY.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = scheduleAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntitySettings:
		entities := make([]settingsTY.Settings, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = settingsAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	case types.EntityDataRepository:
		entities := make([]dataRepositoryTY.Config, 0)
		err := unmarshal(fileFormat, fileBytes, &entities)
		if err != nil {
			return err
		}
		for index := 0; index < len(entities); index++ {
			err = dataRepositoryAPI.Save(&entities[index])
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown entity type:%s", entityName)
	}

	return nil
}

func unmarshal(provider string, fileBytes []byte, entities interface{}) error {
	switch provider {
	case backupTY.TypeJSON:
		err := json.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	case backupTY.TypeYAML:
		err := yaml.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown provider:%s", provider)
	}
	return nil
}

func getEntityName(filename string) string {
	entity := strings.Split(filename, backupTY.EntityNameIndexSplit)
	if len(entity) > 0 {
		return entity[0]
	}
	return ""
}

func ExtractExportedZipfile(exportedZipfile string) error {
	if isImportJobRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isImportJobRunning.Set()
	defer isImportJobRunning.Reset()

	zipFilename := path.Base(exportedZipfile)
	baseDir := strings.TrimSuffix(zipFilename, path.Ext(zipFilename))
	extractFullPath := path.Join(types.GetDataDirectoryInternal(), baseDir)

	err := ziputils.Unzip(exportedZipfile, extractFullPath)
	if err != nil {
		zap.L().Error("error on unzip", zap.String("exportedZipfile", exportedZipfile), zap.String("extractLocation", extractFullPath), zap.Error(err))
		return err
	}

	// TODO: verify extracted file

	systemStartJobs := &config.SystemStartupJobs{
		Restore: config.StartupRestore{
			Enabled:            true,
			ExtractedDirectory: extractFullPath,
			ClearDatabase:      true,
		},
	}

	// store on internal, will be restored on startup
	dataBytes, err := yaml.Marshal(systemStartJobs)
	if err != nil {
		return err
	}

	internalDir := types.GetDataDirectoryInternal()
	err = utils.WriteFile(internalDir, config.SystemStartJobsFilename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", internalDir), zap.String("filename", config.SystemStartJobsFilename), zap.Error(err))
		return err
	}

	return nil
}
