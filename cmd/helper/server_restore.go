package helper

import (
	"fmt"

	bkpMap "github.com/mycontroller-org/server/v2/pkg/backup"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storage "github.com/mycontroller-org/server/v2/plugin/database/storage"
	backupAPI "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// returns restoreEngine api and storage import api map
// used in data import and restore
func (s *Server) getRestoreMap() (*backupAPI.BackupRestore, map[string]backupAPI.Backup, error) {
	bkpDirs, err := bkpMap.GetDirectories(true, true)
	if err != nil {
		return nil, nil, err
	}

	restoreEngine, err := backupAPI.New(s.ctx, bkpDirs)
	if err != nil {
		s.logger.Error("error on getting backup/restore api", zap.Error(err))
		return nil, nil, err

	}
	storageApiMap, err := bkpMap.GetStorageApiMap(s.ctx)
	if err != nil {
		s.logger.Error("error on getting storage api map", zap.Error(err))
		return nil, nil, err
	}

	return restoreEngine, storageApiMap, nil
}

// loads data from disk to database
// for now used only in memory database
func (s *Server) runStorageImport() error {
	restoreEngine, storageApiMap, err := s.getRestoreMap()
	if err != nil {
		return err
	}

	err = storage.RunImport(s.ctx, s.logger, s.storage, storageApiMap, restoreEngine.ExecuteImportStorage)
	if err != nil {
		s.logger.Error("error on storage import", zap.Error(err))
		return err
	}
	return nil
}

// performs restore operation
// verifies on startup and if restore required, performs the operation
func (s *Server) checkSystemRestore() {
	dirInternal := s.config.Directories.GetDataInternal()
	filename := fmt.Sprintf("%s/%s", dirInternal, config.SystemStartJobsFilename)
	if !utils.IsFileExists(filename) {
		return
	}

	bytes, err := utils.ReadFile(dirInternal, config.SystemStartJobsFilename)
	if err != nil {
		s.logger.Fatal("error on loading system startup file", zap.String("filename", filename), zap.Error(err))
		return
	}

	jobs := &config.SystemStartupJobs{}
	err = yaml.Unmarshal(bytes, jobs)
	if err != nil {
		s.logger.Fatal("error on loading system startup file", zap.String("filename", filename), zap.Error(err))
		return
	}

	// perform restore operation
	s.triggerSystemRestore(&jobs.Restore)

	// remove the directories
	err = utils.RemoveFileOrEmptyDir(filename)
	if err != nil {
		s.logger.Fatal("error on removing file", zap.Any("filename", filename), zap.Error(err))
		return
	}

	// remove internal directory
	err = utils.RemoveDir(dirInternal)
	if err != nil {
		s.logger.Fatal("error on removing internal directory on restore operation", zap.String("path", dirInternal), zap.Error(err))
	}
}

// performs restore operation
func (s *Server) triggerSystemRestore(cfg *config.StartupRestore) {
	if !cfg.Enabled {
		return
	}

	restoreEngine, storageApiMap, err := s.getRestoreMap()
	if err != nil {
		s.logger.Fatal("error on getting restore engine and storage api map", zap.Error(err))
		return
	}

	s.logger.Info("found a restore setup on startup. Performing restore operation", zap.Any("config", cfg))

	err = restoreEngine.ExecuteRestore(s.storage, storageApiMap, cfg.ExtractedDirectory)
	if err != nil {
		s.logger.Fatal("error on restore", zap.Error(err))
		return
	}

	// clean extracted files
	err = utils.RemoveDir(cfg.ExtractedDirectory)
	if err != nil {
		s.logger.Fatal("error on deleting extracted backup files", zap.Any("restoreConfig", cfg), zap.Error(err))
		return
	}
}
