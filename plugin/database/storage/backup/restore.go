package backup

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/server/v2/pkg/utils/ziputils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// we can not allowed more than one restore operation at a time
// hence defined it globally
var (
	isRestoreRunning = concurrency.SafeBool{}
)

func (br *BackupRestore) ExecuteRestore(storage storageTY.Plugin, apiMap map[string]Backup, extractedDir string) error {
	start := time.Now()
	br.logger.Info("restore job triggered", zap.String("extractedDirectory", extractedDir))

	err := storage.Pause()
	if err != nil {
		br.logger.Fatal("error on pause a database", zap.Error(err))
		return err
	}

	err = storage.ClearDatabase()
	if err != nil {
		br.logger.Fatal("error on emptying database", zap.Error(err))
		return err
	}

	storageDir := path.Join(extractedDir, StorageBackupDirectoryName)

	dataBytes, err := utils.ReadFile(extractedDir, BackupDetailsFilename)
	if err != nil {
		br.logger.Fatal("error on reading export details", zap.String("dir", extractedDir), zap.String("filename", BackupDetailsFilename), zap.Error(err))
		return err
	}

	exportDetails := &BackupDetails{}
	err = yaml.Unmarshal(dataBytes, exportDetails)
	if err != nil {
		br.logger.Fatal("error on loading export details", zap.Error(err))
		return err
	}

	err = br.ExecuteImportStorage(apiMap, storageDir, exportDetails.StorageExportType, false)
	if err != nil {
		br.logger.Fatal("error on importing storage files", zap.Error(err))
		return err
	}

	err = storage.Resume()
	if err != nil {
		br.logger.Fatal("error on resume a database service", zap.Error(err))
		return err
	}
	br.logger.Info("import database completed", zap.String("timeTaken", time.Since(start).String()))

	// restore directories
	br.logger.Info("restore directories started")
	err = br.restoreDirectories(extractedDir, br.directories)
	if err != nil {
		return err
	}

	br.logger.Info("restore directories successfully completed", zap.String("timeTaken", time.Since(start).String()))
	return nil
}

// restores the secure and insecure shares, if available in the backup
// overwrites if the file exists on the destination directory
func (br *BackupRestore) restoreDirectories(extractedBaseDir string, directories map[string]string) error {
	for name, dstDirFullPath := range directories {
		if name == StorageBackupDirectoryName {
			return fmt.Errorf("name '%s' is reserved for storage database backup", StorageBackupDirectoryName)
		}

		// copy a directory if available in the backup
		sourceDirFullPath := path.Join(extractedBaseDir, name)
		br.logger.Debug("restoring a directory", zap.String("source", sourceDirFullPath), zap.String("dst", dstDirFullPath))
		if utils.IsDirExists(sourceDirFullPath) {
			err := CopyFiles(br.logger, sourceDirFullPath, dstDirFullPath, true)
			if err != nil {
				br.logger.Error("error on coping a directory", zap.String("name", name), zap.String("path", dstDirFullPath), zap.String("source", sourceDirFullPath), zap.String("destination", dstDirFullPath))
				return err
			}
		}
	}
	return nil
}

// ExecuteImportStorage update data into database
func (br *BackupRestore) ExecuteImportStorage(apiMap map[string]Backup, sourceDir, fileType string, ignoreEmptyDir bool) error {
	if isRestoreRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isRestoreRunning.Set()
	defer func() {
		isRestoreRunning.Reset()
		// resume bus service
		br.bus.ResumePublish()
	}()

	// pause publish service in bus
	br.bus.PausePublish()

	br.logger.Info("executing import job", zap.String("sourceDir", sourceDir), zap.String("fileType", fileType))
	// check directory availability
	if !utils.IsDirExists(sourceDir) {
		return fmt.Errorf("specified directory not available. sourceDir:%s", sourceDir)
	}
	files, err := utils.ListFiles(sourceDir)
	if err != nil {
		br.logger.Error("error on listing files", zap.Error(err))
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
		entityName := br.getEntityName(file.Name)
		br.logger.Debug("importing a file", zap.String("file", file.Name), zap.String("entityName", entityName))
		// read data from file
		fileBytes, err := utils.ReadFile(sourceDir, file.Name)
		if err != nil {
			br.logger.Error("error on reading a file", zap.String("fileName", file.FullPath), zap.Error(err))
			return err
		}
		apiHolder, found := apiMap[entityName]
		if !found {
			br.logger.Error("error on getting api map details", zap.String("entityName", entityName))
			return fmt.Errorf("error on getting api map details. entityName:%s", entityName)
		}

		err = br.updateEntities(apiHolder, fileBytes, fileType)
		if err != nil {
			br.logger.Error("error on updating entity", zap.String("file", file.FullPath), zap.String("entityName", entityName), zap.String("fileType", fileType), zap.Error(err))
			return err
		}
	}
	return nil
}

func (br *BackupRestore) updateEntities(api Backup, fileBytes []byte, fileFormat string) error {
	// get actual type
	entityType := reflect.TypeOf(api.GetEntityInterface())
	entities := reflect.New(reflect.SliceOf(entityType)).Interface()

	// load file data into entities
	err := br.unmarshal(fileFormat, fileBytes, entities)
	if err != nil {
		return err
	}

	entitiesElem := reflect.ValueOf(entities).Elem()
	// store the entities to storage database
	for index := 0; index < entitiesElem.Len(); index++ {
		err = api.Import(entitiesElem.Index(index).Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

func (br *BackupRestore) unmarshal(provider string, fileBytes []byte, entities interface{}) error {
	switch provider {
	case TypeJSON:
		err := json.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	case TypeYAML:
		err := yaml.Unmarshal(fileBytes, entities)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown provider:%s", provider)
	}
	return nil
}

func (br *BackupRestore) getEntityName(filename string) string {
	entity := strings.Split(filename, EntityNameIndexSplit)
	if len(entity) > 0 {
		return entity[0]
	}
	return ""
}

func (br *BackupRestore) ExtractExportedZipFile(exportedZipFile, targetDir, restoreReferenceDir, restoreReferenceFilename string) error {
	if isRestoreRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isRestoreRunning.Set()
	defer isRestoreRunning.Reset()

	zipFilename := path.Base(exportedZipFile)
	baseDir := strings.TrimSuffix(zipFilename, path.Ext(zipFilename))
	extractFullPath := path.Join(targetDir, baseDir)

	err := ziputils.Unzip(exportedZipFile, extractFullPath)
	if err != nil {
		br.logger.Error("error on unzip", zap.String("exportedZipfile", exportedZipFile), zap.String("extractLocation", extractFullPath), zap.Error(err))
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

	err = utils.WriteFile(restoreReferenceDir, restoreReferenceFilename, dataBytes)
	if err != nil {
		br.logger.Error("failed to write data to disk", zap.String("directory", restoreReferenceDir), zap.String("filename", restoreReferenceFilename), zap.Error(err))
		return err
	}

	return nil
}
