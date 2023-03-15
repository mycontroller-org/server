package export

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"github.com/mycontroller-org/server/v2/pkg/version"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// do not allow more than one backup job at a time, may lead system performance issue
// hence defined globally
var (
	isBackupRunning = concurrency.SafeBool{}
)

type BackupRestore struct {
	ctx             context.Context
	logger          *zap.Logger
	storage         storageTY.Plugin
	bus             busTY.Plugin
	dirDataInternal string
	dirDataFirmware string
}

func New(ctx context.Context) (*BackupRestore, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	storage, err := storageTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// set dirs
	dirDataInternal := types.GetEnvString(types.ENV_DIR_DATA_INTERNAL)
	if dirDataInternal == "" {
		return nil, fmt.Errorf("environment '%s' not set", types.ENV_DIR_DATA_INTERNAL)
	}
	dirDataFirmware := types.GetEnvString(types.ENV_DIR_DATA_FIRMWARE)
	if dirDataFirmware == "" {
		return nil, fmt.Errorf("environment '%s' not set", types.ENV_DIR_DATA_FIRMWARE)
	}

	return &BackupRestore{
		ctx:             ctx,
		logger:          logger,
		storage:         storage,
		bus:             bus,
		dirDataInternal: dirDataInternal,
		dirDataFirmware: dirDataFirmware,
	}, nil
}

// exports data from database to disk
func (br *BackupRestore) ExportStorage(exportMap map[string]backupTY.Backup, targetDir, exportFormat string) error {
	if isBackupRunning.IsSet() {
		return errors.New("there is a exporter job in progress")
	}
	isBackupRunning.Set()
	defer isBackupRunning.Reset()

	// include version details
	err := br.addBackupInformation(targetDir, exportFormat)
	if err != nil {
		return err
	}

	targetDirFullPath := fmt.Sprintf("%s%s", targetDir, config.DirectoryDataStorage)

	for entityName := range exportMap {
		storageApi := exportMap[entityName]
		p := &storageTY.Pagination{
			Limit: backupTY.LimitPerFile, SortBy: []storageTY.Sort{{Field: types.KeyFieldID, OrderBy: "asc"}}, Offset: 0,
		}
		offset := int64(0)
		for {
			p.Offset = offset
			result, err := storageApi.List(nil, p)
			if err != nil {
				br.logger.Error("failed to get entities", zap.String("entityName", entityName), zap.Error(err))
				return err
			}

			br.exportEntity(targetDirFullPath, entityName, int(result.Offset), result.Data, exportFormat)

			offset += backupTY.LimitPerFile
			if result.Count < offset {
				break
			}
		}
	}
	return nil
}

// updates backup information
func (br *BackupRestore) addBackupInformation(targetDir, storageExportType string) error {
	backupDetails := &backupTY.BackupDetails{
		Filename:          path.Base(targetDir),
		StorageExportType: storageExportType,
		CreatedOn:         time.Now(),
		Version:           version.Get(),
	}

	dataBytes, err := yaml.Marshal(backupDetails)
	if err != nil {
		return err
	}

	err = utils.WriteFile(targetDir, backupTY.BackupDetailsFilename, dataBytes)
	if err != nil {
		br.logger.Error("failed to write data to disk", zap.String("directory", targetDir), zap.String("filename", backupTY.BackupDetailsFilename), zap.Error(err))
		return err
	}
	return nil
}

func (br *BackupRestore) exportEntity(targetDir, entityName string, index int, data interface{}, storageExportType string) {
	// update "User" to "UserWithPassword" to keep the password in json export
	if entityName == types.EntityUser {
		if users, ok := data.(*[]userTY.User); ok {
			usersWithPasswd := make([]userTY.UserWithPassword, len(*users))
			for _index, user := range *users {
				usersWithPasswd[_index] = userTY.UserWithPassword(user)
			}
			if len(usersWithPasswd) > 0 {
				data = usersWithPasswd
			}
		} else {
			br.logger.Error("error on converting the data to user slice, continue with default data type", zap.String("inputType", fmt.Sprintf("%T", data)))
		}
	}

	var dataBytes []byte
	var err error
	switch storageExportType {
	case backupTY.TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			br.logger.Error("failed to convert to target format", zap.String("format", storageExportType), zap.Error(err))
			return
		}
	case backupTY.TypeYAML:
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			br.logger.Error("failed to convert to target format", zap.String("format", storageExportType), zap.Error(err))
			return
		}
	default:
		br.logger.Error("This format not supported", zap.String("format", storageExportType), zap.Error(err))
		return
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, "__", index, storageExportType)
	dir := fmt.Sprintf("%s/%s", targetDir, storageExportType)
	err = utils.WriteFile(targetDir, filename, dataBytes)
	if err != nil {
		br.logger.Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}

// copies files from a location to another location
func (br *BackupRestore) CopyFiles(sourceDir, dstDir string, overwrite bool) error {
	err := utils.CreateDir(dstDir)
	if err != nil {
		return err
	}

	files, err := utils.ListFiles(sourceDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		destPath := fmt.Sprintf("%s/%s", dstDir, file.Name)
		err = utils.CopyFileForce(file.FullPath, destPath, overwrite)
		if err != nil {
			return err
		}
	}

	return nil
}
