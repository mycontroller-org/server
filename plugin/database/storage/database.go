package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type updateRestoreApiMap func(storage storageTY.Plugin, backupVersion string, apiMap map[string]backupTY.Backup) (map[string]backupTY.Backup, error)

// Init storage service
func Get(ctx context.Context, storageCfg cmap.CustomMap) (storageTY.Plugin, error) {
	// get plugin type
	pluginType := storageCfg.GetString(storageTY.KeyType)
	if pluginType == "" {
		return nil, errors.New("error on storage database initialization, type not defined")
	}

	plugin, err := Create(ctx, pluginType, storageCfg)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}

// should be run at startup
// in-memory database will be empty at startup time
// this function checks the "storage.DoStartupImport" and performs the operation, if the database supports
func RunImport(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, apiMap map[string]backupTY.Backup, importFunc func(apiMap map[string]backupTY.Backup, targetDir, fileType string, ignoreEmptyDir bool) error, updateRestoreApiMapFn updateRestoreApiMap) error {
	_logger := logger.Named("run_storage_import")
	if doStartImport, filesDir, fileFormat := storage.DoStartupImport(); doStartImport {
		_logger.Debug("startup import requested")
		err := utils.CreateDir(filesDir)
		if err != nil {
			_logger.Fatal("error on creating files directory", zap.Error(err), zap.String("filesDir", filesDir))
			return err
		}

		versionSettings, err := getVersionFromFile(logger, filesDir, fileFormat)
		if err != nil {
			return err
		}
		// update restore api with actual backed up server version
		if versionSettings != nil {
			_updateApiMap, err := updateRestoreApiMapFn(storage, versionSettings.Version, apiMap)
			if err != nil {
				return err
			}
			apiMap = _updateApiMap
		}

		err = importFunc(apiMap, filesDir, fileFormat, true)
		if err != nil {
			_logger.Fatal("error on run startup import on database", zap.Error(err))
			// _logger.WithOptions(zap.AddCallerSkip(10)).Error("error on local import", zap.String("error", err.Error()))
			return err
		}
	}

	err := storage.Resume()
	if err != nil {
		_logger.Fatal("error on resuming the database service", zap.Error(err))
	}
	return nil
}

func getVersionFromFile(_logger *zap.Logger, filesDir, fileFormat string) (*settingsTY.VersionSettings, error) {
	files, err := utils.ListFiles(filesDir)
	if err != nil {
		_logger.Error("unable to list the files on the given location", zap.String("dir", filesDir), zap.Error(err))
		return nil, err
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name, "settings") {
			_logger.Debug("settings file found", zap.String("fileDetails", file.ToString()))
			// read the file
			fileData, err := utils.ReadFile(filesDir, file.Name)
			if err != nil {
				_logger.Error("error on reading a file", zap.String("filename", file.FullPath), zap.Error(err))
				return nil, err
			}
			// convert to settings spec
			settings := []settingsTY.Settings{}
			switch fileFormat {
			case backupTY.TypeJSON:
				err = json.Unmarshal(fileData, &settings)
				if err != nil {
					_logger.Error("error on unmarshal a file with json", zap.String("filename", file.FullPath), zap.Error(err))
					return nil, err
				}
			case backupTY.TypeYAML:
				err = yaml.Unmarshal(fileData, &settings)
				if err != nil {
					_logger.Error("error on unmarshal a file with yaml", zap.String("filename", file.FullPath), zap.Error(err))
					return nil, err
				}
			default:
				return nil, fmt.Errorf("received unknown fileFormat:%s", fileFormat)
			}
			for _, setting := range settings {
				if setting.ID == settingsTY.KeyVersion {
					versionSettings := settingsTY.VersionSettings{}
					err = utils.MapToStruct(utils.TagNameNone, setting.Spec, &versionSettings)
					if err != nil {
						_logger.Error("error on map to struct", zap.String("filename", file.FullPath), zap.Error(err))
						return nil, err
					}
					_logger.Info("found version details", zap.Any("version", versionSettings))
					return &versionSettings, nil
				}
			}
		}
	}

	_logger.Info("unable to get version details from disk, this could be a very first run")
	return nil, nil
}
