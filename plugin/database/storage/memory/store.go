package memory

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	PluginMemory = storageTY.TypeMemory

	loggerName = "memory_db"

	defaultSyncInterval = "1m"
	syncJobName         = "in-memory-db-sync-to-disk"

	defaultDumpDir    = "memory_db"
	defaultDumpFormat = backupTY.TypeJSON
)

// Config of the memory storage
type Config struct {
	Name         string   `yaml:"name"`
	DumpEnabled  bool     `yaml:"dump_enabled"`
	DumpInterval string   `yaml:"dump_interval"`
	DumpDir      string   `yaml:"dump_dir"`
	DumpFormat   []string `yaml:"dump_format"`
	LoadFormat   string   `yaml:"load_format"`
}

// Store to keep all the entities
type Store struct {
	mutex     *sync.RWMutex
	Config    Config
	data      map[string][]interface{} // entities map with entity name
	lastSync  time.Time
	paused    bool
	logger    *zap.Logger
	scheduler schedulerTY.CoreScheduler
}

// New in-memory database
func New(ctx context.Context, config cmap.CustomMap) (storageTY.Plugin, error) {
	logger := storageTY.GetStorageLogger()

	// get required plugins
	scheduler, err := schedulerTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	err = utils.MapToStruct(utils.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.DumpDir == "" {
		cfg.DumpDir = cfg.Name
	}

	if cfg.LoadFormat == "" {
		cfg.LoadFormat = defaultDumpFormat
	}

	// update default dump format, if none supplied
	if len(cfg.DumpFormat) == 0 {
		cfg.DumpFormat = []string{defaultDumpFormat}
	}

	store := &Store{
		Config:    cfg,
		data:      make(map[string][]interface{}),
		lastSync:  time.Now(),
		mutex:     &sync.RWMutex{},
		logger:    logger.Named(loggerName),
		scheduler: scheduler,
	}

	return store, nil
}

func (s *Store) Name() string {
	return PluginMemory
}

// DoStartupImport returns the needs, files location, and file format
func (s *Store) DoStartupImport() (bool, string, string) {
	return true, s.getStorageLocation(s.Config.LoadFormat), s.Config.LoadFormat
}

func (s *Store) loadDumpJob() error {
	if s.Config.DumpEnabled {
		if s.Config.DumpDir == "" {
			s.Config.DumpDir = defaultDumpDir
		}

		// add sync job
		if s.Config.DumpInterval == "" {
			s.Config.DumpInterval = defaultSyncInterval
		}
		err := s.scheduler.AddFunc(syncJobName, fmt.Sprintf("@every %s", s.Config.DumpInterval), s.writeToDisk)
		if err != nil {
			return err
		}
		s.logger.Debug("Memory database dump job scheduled", zap.Any("config", s.Config))
	}
	return nil
}

// Pause the storage to perform import like jobs
func (s *Store) Pause() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.paused = true
	// stop dump job
	s.scheduler.RemoveFunc(syncJobName)

	return nil
}

// Resume the storage if Paused
func (s *Store) Resume() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.paused = false
	return s.loadDumpJob()
}

// ClearDatabase removes all the data from the database
func (s *Store) ClearDatabase() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// remove all the data
	s.data = make(map[string][]interface{})

	storageDirPath := s.getStorageLocation("")
	s.logger.Info("removing storage directory", zap.String("dir", storageDirPath))
	// remove all the files from disk
	return utils.RemoveDir(storageDirPath)
}

func (s *Store) writeToDisk() {
	start := time.Now()
	s.logger.Debug("data dump to disk job is triggered")
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for entityName, data := range s.data {
		for _, format := range s.Config.DumpFormat {
			itemsCount := len(data)
			if itemsCount == 0 {
				continue
			}
			index := 0
			for {
				positionStart := index * backupTY.LimitPerFile
				index++
				positionEnd := (index * backupTY.LimitPerFile)

				if positionEnd < itemsCount {
					s.dump(entityName, index, data[positionStart:positionEnd], format)
				} else {
					s.dump(entityName, index, data[positionStart:], format)
					break
				}
			}
		}
	}
	s.logger.Debug("data dump to disk job is completed", zap.String("timeTaken", time.Since(start).String()))
}

func (s *Store) dump(entityName string, index int, data interface{}, extension string) {
	// update user to userPassword to keep the password on the json export
	if entityName == types.EntityUser {
		if users, ok := data.([]interface{}); ok {
			usersWithPasswd := make([]userTY.UserWithPassword, 0)
			for _, userInterface := range users {
				user, ok := userInterface.(*userTY.User)
				if !ok {
					s.logger.Error("error on converting the data to user slice, continue with default data type", zap.String("inputType", fmt.Sprintf("%T", userInterface)))
					break
				}
				usersWithPasswd = append(usersWithPasswd, userTY.UserWithPassword(*user))
			}
			if len(usersWithPasswd) > 0 {
				data = usersWithPasswd
			}
		} else {
			s.logger.Error("error on converting the data to user slice, continue with default data type", zap.String("inputType", fmt.Sprintf("%T", data)))
		}
	}

	var dataBytes []byte
	var err error
	switch extension {
	case backupTY.TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			s.logger.Error("failed to convert to target extension", zap.String("extension", extension), zap.Error(err))
			return
		}
	case backupTY.TypeYAML:
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			s.logger.Error("failed to convert to target extension", zap.String("extension", extension), zap.Error(err))
			return
		}

	default:
		s.logger.Error("This extension not supported", zap.String("extension", extension), zap.Error(err))
		return
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, backupTY.EntityNameIndexSplit, index, extension)
	dir := s.getStorageLocation(extension)
	err = utils.WriteFile(dir, filename, dataBytes)
	if err != nil {
		s.logger.Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}

func (s *Store) getStorageLocation(provider string) string {
	storageDir := path.Join(types.GetEnvString(types.ENV_DIR_DATA_STORAGE), s.Config.DumpDir)
	if provider != "" {
		return path.Join(storageDir, provider)
	}
	return storageDir
}
