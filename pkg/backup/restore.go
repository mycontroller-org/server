package export

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	dashboardAPI "github.com/mycontroller-org/backend/v2/pkg/api/dashboard"
	dataRepositoryAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	fwAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	fwdPayloadAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	gwAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	notificationHandlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/handler"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	scheduleAPI "github.com/mycontroller-org/backend/v2/pkg/api/schedule"
	settingsAPI "github.com/mycontroller-org/backend/v2/pkg/api/settings"
	sourceAPI "github.com/mycontroller-org/backend/v2/pkg/api/source"
	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	backupML "github.com/mycontroller-org/backend/v2/pkg/model/backup"
	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	dashboardML "github.com/mycontroller-org/backend/v2/pkg/model/dashboard"
	dataRepositoryML "github.com/mycontroller-org/backend/v2/pkg/model/data_repository"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	firmwareML "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	fwdPayloadML "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	gatewayML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	nhML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	scheduleML "github.com/mycontroller-org/backend/v2/pkg/model/schedule"
	settingsML "github.com/mycontroller-org/backend/v2/pkg/model/settings"
	sourceML "github.com/mycontroller-org/backend/v2/pkg/model/source"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/backend/v2/pkg/utils/ziputils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var isImportJobRunning = concurrency.SafeBool{}

func ExecuteRestore(extractedDir string) error {
	start := time.Now()
	zap.L().Info("Restore job triggered", zap.String("extractedDirectory", extractedDir))

	err := storage.SVC.Pause()
	if err != nil {
		zap.L().Fatal("error on pause a database", zap.Error(err))
		return err
	}

	err = storage.SVC.ClearDatabase()
	if err != nil {
		zap.L().Fatal("error on emptying database", zap.Error(err))
		return err
	}

	storageDir := path.Join(extractedDir, model.DirectoryDataStorage)
	firmwareDir := path.Join(extractedDir, model.DirectoryDataFirmware)

	dataBytes, err := utils.ReadFile(extractedDir, backupML.BackupDetailsFilename)
	if err != nil {
		zap.L().Fatal("error on reading export details", zap.String("dir", extractedDir), zap.String("filename", backupML.BackupDetailsFilename), zap.Error(err))
		return err
	}

	exportDetails := &backupML.BackupDetails{}
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

	err = storage.SVC.Resume()
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

	destDir := model.GetDataDirectoryFirmware()
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
			return err
		}
		err = updateEntities(fileBytes, entityName, fileType)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateEntities(fileBytes []byte, entityName, fileFormat string) error {
	switch entityName {
	case model.EntityGateway:
		entities := make([]gatewayML.Config, 0)
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

	case model.EntityNode:
		entities := make([]nodeML.Node, 0)
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

	case model.EntitySource:
		entities := make([]sourceML.Source, 0)
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

	case model.EntityField:
		entities := make([]fieldML.Field, 0)
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

	case model.EntityFirmware:
		entities := make([]firmwareML.Firmware, 0)
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

	case model.EntityUser:
		entities := make([]userML.User, 0)
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

	case model.EntityDashboard:
		entities := make([]dashboardML.Config, 0)
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

	case model.EntityHandler:
		entities := make([]nhML.Config, 0)
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

	case model.EntityForwardPayload:
		entities := make([]fwdPayloadML.Config, 0)
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

	case model.EntityTask:
		entities := make([]taskML.Config, 0)
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

	case model.EntitySchedule:
		entities := make([]scheduleML.Config, 0)
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

	case model.EntitySettings:
		entities := make([]settingsML.Settings, 0)
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

	case model.EntityDataRepository:
		entities := make([]dataRepositoryML.Config, 0)
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
	case backupML.TypeJSON:
		err := json.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	case backupML.TypeYAML:
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
	entity := strings.Split(filename, backupML.EntityNameIndexSplit)
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
	extractFullPath := path.Join(model.GetDataDirectoryInternal(), baseDir)

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

	internalDir := model.GetDataDirectoryInternal()
	err = utils.WriteFile(internalDir, config.SystemStartJobsFilename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", internalDir), zap.String("filename", config.SystemStartJobsFilename), zap.Error(err))
		return err
	}

	return nil
}
