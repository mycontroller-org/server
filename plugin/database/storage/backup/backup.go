package backup

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
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
	ctx         context.Context
	logger      *zap.Logger
	storage     storageTY.Plugin
	bus         busTY.Plugin
	directories map[string]string
}

func New(ctx context.Context, directories map[string]string) (*BackupRestore, error) {
	logger, err := loggerUtils.FromContext(ctx)
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

	return &BackupRestore{
		ctx:         ctx,
		logger:      logger,
		storage:     storage,
		bus:         bus,
		directories: directories,
	}, nil
}

// exports data from database to disk
func (br *BackupRestore) ExportStorage(exportMap map[string]Backup, transformerFunc DataTransformerFunc, targetDir, exportFormat string) error {
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

	// export directories
	err = br.exportDirectories(targetDir)
	if err != nil {
		return err
	}

	// export database
	targetDirFullPath := fmt.Sprintf("%s/%s", targetDir, StorageBackupDirectoryName)
	for entityName := range exportMap {
		storageApi := exportMap[entityName]
		p := &storageTY.Pagination{
			Limit: LimitPerFile, SortBy: []storageTY.Sort{{Field: types.KeyFieldID, OrderBy: "asc"}}, Offset: 0,
		}
		offset := int64(0)
		for {
			p.Offset = offset
			result, err := storageApi.List(nil, p)
			if err != nil {
				br.logger.Error("failed to get entities", zap.String("entityName", entityName), zap.Error(err))
				return err
			}

			err = br.exportEntity(targetDirFullPath, entityName, int(result.Offset), result.Data, exportFormat, transformerFunc)
			if err != nil {
				return err
			}

			offset += LimitPerFile
			if result.Count < offset {
				break
			}
		}
	}
	return nil
}

// updates backup information
func (br *BackupRestore) addBackupInformation(targetDir, storageExportType string) error {
	backupDetails := &BackupDetails{
		Filename:          path.Base(targetDir),
		StorageExportType: storageExportType,
		CreatedOn:         time.Now(),
		Version:           version.Get(),
		Directories:       br.directories,
	}

	dataBytes, err := yaml.Marshal(backupDetails)
	if err != nil {
		return err
	}

	err = utils.WriteFile(targetDir, BackupDetailsFilename, dataBytes)
	if err != nil {
		br.logger.Error("failed to write data to disk", zap.String("directory", targetDir), zap.String("filename", BackupDetailsFilename), zap.Error(err))
		return err
	}
	return nil
}

func (br *BackupRestore) exportEntity(targetDir, entityName string, index int, data interface{}, storageExportType string, transformerFunc DataTransformerFunc) error {
	if transformerFunc != nil {
		_data, err := transformerFunc(br.logger, entityName, data, storageExportType)
		if err != nil {
			br.logger.Error("error on executing transformer function", zap.String("entityName", entityName), zap.Error(err))
			return err
		}
		data = _data
	}

	var dataBytes []byte
	var err error
	switch storageExportType {
	case TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			br.logger.Error("failed to convert to target format", zap.String("format", storageExportType), zap.Error(err))
			return err
		}
	case TypeYAML:
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			br.logger.Error("failed to convert to target format", zap.String("format", storageExportType), zap.Error(err))
			return err
		}
	default:
		br.logger.Error("This format not supported", zap.String("format", storageExportType), zap.Error(err))
		return err
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, "__", index, storageExportType)
	dir := fmt.Sprintf("%s/%s", targetDir, storageExportType)
	err = utils.WriteFile(targetDir, filename, dataBytes)
	if err != nil {
		br.logger.Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
		return err
	}
	return nil
}

func (br *BackupRestore) exportDirectories(targetDir string) error {
	for name, path := range br.directories {
		if name == StorageBackupDirectoryName {
			return fmt.Errorf("name '%s' is reserved for storage database backup", StorageBackupDirectoryName)
		}
		targetDirFullPath := fmt.Sprintf("%s/%s", targetDir, name)
		err := CopyFiles(br.logger, path, targetDirFullPath, false)
		if err != nil {
			br.logger.Error("error on copying a directory", zap.String("name", name), zap.String("path", path), zap.Error(err))
			return err
		}
	}
	return nil
}
