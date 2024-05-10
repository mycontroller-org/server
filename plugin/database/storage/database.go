package storage

import (
	"context"
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

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
func RunImport(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, apiMap map[string]backupTY.Backup, importFunc func(apiMap map[string]backupTY.Backup, targetDir, fileType string, ignoreEmptyDir bool) error) error {
	_logger := logger.Named("run_storage_import")
	if doStartImport, filesDir, fileFormat := storage.DoStartupImport(); doStartImport {
		_logger.Debug("startup import requested")
		err := utils.CreateDir(filesDir)
		if err != nil {
			_logger.Fatal("error on creating files directory", zap.Error(err), zap.String("filesDir", filesDir))
			return err
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
