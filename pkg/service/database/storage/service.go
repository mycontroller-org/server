package storage

import (
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	cfgML "github.com/mycontroller-org/server/v2/pkg/model/config"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storagePlugin "github.com/mycontroller-org/server/v2/plugin/database/storage"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

// Init storage service
func Init(storageCfg cmap.CustomMap, loggerCfg cfgML.LoggerConfig) (stgType.Plugin, error) {
	// include logger details
	storageCfg["logger"] = map[string]string{"mode": loggerCfg.Mode, "encoding": loggerCfg.Encoding, "level": loggerCfg.Level.Storage}

	// get plugin type
	pluginType := storageCfg.GetString(model.KeyType)
	if pluginType == "" {
		return nil, errors.New("error on storage database initialization, type not defined")
	}

	plugin, err := storagePlugin.Create(pluginType, storageCfg)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}

func RunImport(plugin stgType.Plugin, importFunc func(targetDir, fileType string, ignoreEmptyDir bool) error) error {
	if doStartImport, filesDir, fileFormat := plugin.DoStartupImport(); doStartImport {
		// run startup import
		// Pause Timestamp Update and resume later
		configuration.PauseModifiedOnUpdate.Set()
		defer configuration.PauseModifiedOnUpdate.Reset()

		zap.L().Debug("startup import requested")
		err := utils.CreateDir(filesDir)
		if err != nil {
			zap.L().Fatal("error on creating files directory", zap.Error(err), zap.String("filesDir", filesDir))
			return err
		}
		err = importFunc(filesDir, fileFormat, true)
		if err != nil {
			zap.L().Fatal("error on run startup import on database", zap.Error(err))
			// zap.L().WithOptions(zap.AddCallerSkip(10)).Error("error on local import", zap.String("error", err.Error()))
			return err
		}
	}

	err := plugin.Resume()
	if err != nil {
		zap.L().Fatal("error on resuming the database service", zap.Error(err))
		return err
	}
	return nil
}
