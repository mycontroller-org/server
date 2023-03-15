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
	"github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
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

func (br *BackupRestore) ExecuteRestore(storage storageTY.Plugin, apiMap map[string]backupTY.Backup, extractedDir string) error {
	start := time.Now()
	br.logger.Info("Restore job triggered", zap.String("extractedDirectory", extractedDir))

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

	storageDir := path.Join(extractedDir, config.DirectoryDataStorage)
	firmwareDir := path.Join(extractedDir, config.DirectoryDataFirmware)

	dataBytes, err := utils.ReadFile(extractedDir, backupTY.BackupDetailsFilename)
	if err != nil {
		br.logger.Fatal("error on reading export details", zap.String("dir", extractedDir), zap.String("filename", backupTY.BackupDetailsFilename), zap.Error(err))
		return err
	}

	exportDetails := &backupTY.BackupDetails{}
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

	br.logger.Info("Import database completed", zap.String("timeTaken", time.Since(start).String()))

	// restore firmwares
	err = br.restoreFirmwares(firmwareDir)
	if err != nil {
		br.logger.Fatal("error on copying firmware files", zap.Error(err))
		return err
	}

	// restore secure and insecure shares
	err = br.restoreSecureInsecureShare(extractedDir)
	if err != nil {
		br.logger.Fatal("error on copying secure or insecure shares files", zap.Error(err))
		return err
	}

	br.logger.Info("restore completed successfully", zap.String("timeTaken", time.Since(start).String()))
	return nil
}

// restores the secure and insecure shares, if available in the backup
// overwrites if the file exists on the destination directory
func (br *BackupRestore) restoreSecureInsecureShare(extractedDir string) error {
	// copy secure share directory if available in the backup
	secureShareDir := path.Join(extractedDir, config.DirectorySecureShare)
	if utils.IsDirExists(secureShareDir) {
		secureShareDst := types.GetEnvString(types.ENV_DIR_SHARE_SECURE)
		if secureShareDst == "" {
			return fmt.Errorf("environment '%s' not set", types.ENV_DIR_SHARE_SECURE)
		}
		err := br.CopyFiles(secureShareDir, secureShareDst, true)
		if err != nil {
			br.logger.Error("error on coping the secure share files", zap.String("source", secureShareDir), zap.String("destination", secureShareDst))
			return err
		}
	}

	// copy insecure share directory if available in the backup
	insecureShareDir := path.Join(extractedDir, config.DirectoryInsecureShare)
	if utils.IsDirExists(insecureShareDir) {
		insecureShareDst := types.GetEnvString(types.ENV_DIR_SHARE_INSECURE)
		if insecureShareDst == "" {
			return fmt.Errorf("environment '%s' not set", types.ENV_DIR_SHARE_INSECURE)
		}
		err := br.CopyFiles(insecureShareDir, insecureShareDst, true)
		if err != nil {
			br.logger.Error("error on coping the insecure share files", zap.String("source", insecureShareDir), zap.String("destination", insecureShareDst))
			return err
		}
	}

	return nil

}

// restores firmwares to the actual directory
func (br *BackupRestore) restoreFirmwares(sourceDir string) error {
	if isRestoreRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isRestoreRunning.Set()
	defer isRestoreRunning.Reset()

	destDir := br.dirDataFirmware
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
func (br *BackupRestore) ExecuteImportStorage(apiMap map[string]backupTY.Backup, sourceDir, fileType string, ignoreEmptyDir bool) error {
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

func (br *BackupRestore) updateEntities(api backupTY.Backup, fileBytes []byte, fileFormat string) error {
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

func (br *BackupRestore) getEntityName(filename string) string {
	entity := strings.Split(filename, backupTY.EntityNameIndexSplit)
	if len(entity) > 0 {
		return entity[0]
	}
	return ""
}

func (br *BackupRestore) ExtractExportedZipFile(exportedZipFile string) error {
	if isRestoreRunning.IsSet() {
		return errors.New("there is an import job is in progress")
	}
	isRestoreRunning.Set()
	defer isRestoreRunning.Reset()

	zipFilename := path.Base(exportedZipFile)
	baseDir := strings.TrimSuffix(zipFilename, path.Ext(zipFilename))
	extractFullPath := path.Join(br.dirDataInternal, baseDir)

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

	internalDir := types.GetEnvString(types.ENV_DIR_DATA_INTERNAL)
	err = utils.WriteFile(internalDir, config.SystemStartJobsFilename, dataBytes)
	if err != nil {
		br.logger.Error("failed to write data to disk", zap.String("directory", internalDir), zap.String("filename", config.SystemStartJobsFilename), zap.Error(err))
		return err
	}

	return nil
}
