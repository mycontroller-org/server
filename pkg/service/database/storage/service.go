package storage

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storagePlugin "github.com/mycontroller-org/server/v2/plugin/database/storage"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

// storage service
var (
	SVC stgType.Plugin
)

// Init storage service
func Init(storageCfg cmap.CustomMap, importFunc func(targetDir, fileType string, ignoreEmptyDir bool) error) {
	// include logger details
	storageCfg["logger"] = map[string]string{"mode": cfg.CFG.Logger.Mode, "encoding": cfg.CFG.Logger.Encoding, "level": cfg.CFG.Logger.Level.Storage}

	// get plugin type
	pluginType := storageCfg.GetString(model.KeyType)
	if pluginType == "" {
		zap.L().Fatal("error on storage database initialization, type not defined")
		return
	}

	plugin, err := storagePlugin.Create(pluginType, storageCfg)
	if err != nil {
		zap.L().Fatal("error on storage database initialization", zap.Error(err))
		return
	}

	// assign the received plugin to storage SVC
	SVC = plugin

	if doStartImport, filesDir, fileFormat := plugin.DoStartupImport(); doStartImport {
		// run startup import
		// Pause Timestamp Update and resume later
		configuration.PauseModifiedOnUpdate.Set()
		defer configuration.PauseModifiedOnUpdate.Reset()

		zap.L().Debug("startup import requested")
		err = utils.CreateDir(filesDir)
		if err != nil {
			zap.L().Fatal("error on creating files directory", zap.Error(err), zap.String("filesDir", filesDir))
			return
		}
		err = importFunc(filesDir, fileFormat, true)
		if err != nil {
			zap.L().Fatal("error on run startup import on database", zap.Error(err))
			// zap.L().WithOptions(zap.AddCallerSkip(10)).Error("error on local import", zap.String("error", err.Error()))
			return
		}
	}

	err = plugin.Resume()
	if err != nil {
		zap.L().Fatal("error on resuming the database service", zap.Error(err))
		return
	}

}
