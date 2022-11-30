package export

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/server/v2/pkg/utils/ziputils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var isImportJobRunning = concurrency.SafeBool{}

func ExecuteRestore(storage storageTY.Plugin, apiMap map[string]backupTY.SaveAPIHolder, extractedDir string) error {
	start := time.Now()
	zap.L().Info("Restore job triggered", zap.String("extractedDirectory", extractedDir))

	err := storage.Pause()
	if err != nil {
		zap.L().Fatal("error on pause a database", zap.Error(err))
		return err
	}

	err = storage.ClearDatabase()
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

	err = ExecuteImportStorage(apiMap, storageDir, exportDetails.StorageExportType, false)
	if err != nil {
		zap.L().Fatal("error on importing storage files", zap.Error(err))
		return err
	}

	err = storage.Resume()
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
func ExecuteImportStorage(apiMap map[string]backupTY.SaveAPIHolder, sourceDir, fileType string, ignoreEmptyDir bool) error {
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

	zap.L().Info("executing import job", zap.String("sourceDir", sourceDir), zap.String("fileType", fileType))
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
		zap.L().Debug("importing a file", zap.String("file", file.Name), zap.String("entityName", entityName))
		// read data from file
		fileBytes, err := utils.ReadFile(sourceDir, file.Name)
		if err != nil {
			zap.L().Error("error on reading a file", zap.String("fileName", file.FullPath), zap.Error(err))
			return err
		}
		apiHolder, found := apiMap[entityName]
		if !found {
			zap.L().Error("error on getting api map details", zap.String("entityName", entityName))
			return fmt.Errorf("error on getting api map details. entityName:%s", entityName)
		}

		err = updateEntities(apiHolder, fileBytes, entityName, fileType)
		if err != nil {
			zap.L().Error("error on updating entity", zap.String("file", file.FullPath), zap.String("entityName", entityName), zap.String("fileType", fileType), zap.Error(err))
			return err
		}
	}
	return nil
}

func updateEntities(apiHolder backupTY.SaveAPIHolder, fileBytes []byte, entityName, fileFormat string) error {
	// get actual type
	entityType := reflect.TypeOf(apiHolder.EntityType)
	entities := reflect.New(reflect.SliceOf(entityType)).Interface()

	// load file data into entities
	err := unmarshal(fileFormat, fileBytes, entities)
	if err != nil {
		return err
	}

	entitiesElem := reflect.ValueOf(entities).Elem()
	// store the entities
	for index := 0; index < entitiesElem.Len(); index++ {
		err = apiHolder.API(entitiesElem.Index(index).Interface())
		if err != nil {
			return err
		}
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

func ExtractExportedZipFile(exportedZipFile string) error {
	if isImportJobRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isImportJobRunning.Set()
	defer isImportJobRunning.Reset()

	zipFilename := path.Base(exportedZipFile)
	baseDir := strings.TrimSuffix(zipFilename, path.Ext(zipFilename))
	extractFullPath := path.Join(types.GetDataDirectoryInternal(), baseDir)

	err := ziputils.Unzip(exportedZipFile, extractFullPath)
	if err != nil {
		zap.L().Error("error on unzip", zap.String("exportedZipfile", exportedZipFile), zap.String("extractLocation", extractFullPath), zap.Error(err))
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
