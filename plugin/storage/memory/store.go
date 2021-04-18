package memory

import (
	"fmt"
	"path"
	"sync"
	"time"

	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	backupML "github.com/mycontroller-org/backend/v2/pkg/model/backup"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	sch "github.com/mycontroller-org/backend/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	defaultSyncInterval = "1m"
	syncJobName         = "in-memory-db-sync-to-disk"

	defaultDumpDir    = "memory_db"
	defaultDumpFormat = backupML.TypeJSON
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
	RWMutex  sync.RWMutex
	Config   Config
	data     map[string][]interface{} // entities map with entity name
	lastSync time.Time
	paused   bool
}

// NewClient in-memory database
func NewClient(config map[string]interface{}) (*Store, error) {
	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameYaml, config, &cfg)
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
		Config:   cfg,
		data:     make(map[string][]interface{}),
		lastSync: time.Now(),
	}

	return store, nil
}

func (s *Store) LocalImport(importFunc func(targetDir, fileType string, ignoreEmptyDir bool) error) error {
	// load data from disk
	dataDir := s.getStorageLocation(s.Config.LoadFormat)
	utils.CreateDir(dataDir)
	err := importFunc(dataDir, s.Config.LoadFormat, true)
	if err != nil {
		zap.L().WithOptions(zap.AddCallerSkip(10)).Error("error on local import", zap.String("error", err.Error()))
		return err
	}

	return s.loadDumpJob()
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
		err := sch.SVC.AddFunc(syncJobName, fmt.Sprintf("@every %s", s.Config.DumpInterval), s.writeToDisk)
		if err != nil {
			return err
		}
		zap.L().Debug("Memory database dump job scheduled", zap.Any("config", s.Config))
	}
	return nil
}

// Pause the storage to perform import like jobs
func (s *Store) Pause() error {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	s.paused = true
	// stop dump job
	sch.SVC.RemoveFunc(syncJobName)

	return nil
}

// Resume the storage if Paused
func (s *Store) Resume() error {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	s.paused = false
	s.loadDumpJob()

	return nil
}

// ClearDatabase removes all the data from the database
func (s *Store) ClearDatabase() error {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	// remove all the data
	s.data = make(map[string][]interface{})

	// remove all the files from disk
	storageDir := path.Join(model.GetDirectoryStorage(), s.Config.DumpDir)
	return utils.RemoveDir(storageDir)
}

func (s *Store) writeToDisk() {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	for entityName, data := range s.data {
		for _, format := range s.Config.DumpFormat {
			dataLength := len(data)
			index := 0
			for {
				if dataLength == 0 {
					break
				}
				positionStart := index * backupML.LimitPerFile
				index++
				positionEnd := (index * backupML.LimitPerFile)

				if positionEnd < dataLength {
					s.dump(entityName, index, data[positionStart:positionEnd-1], format)
				} else {
					s.dump(entityName, index, data[positionStart:], format)
					break
				}
			}
		}
	}
}

func (s *Store) dump(entityName string, index int, data interface{}, provider string) {
	// update user to userPassword to keep the password on the json export
	if entityName == model.EntityUser {
		if users, ok := data.([]interface{}); ok {
			usersWithPasswd := make([]userML.UserWithPassword, 0)
			for _, userInterface := range users {
				user, ok := userInterface.(*userML.User)
				if !ok {
					zap.L().Error("error on converting the data to user slice, continue with default data type", zap.String("inputType", fmt.Sprintf("%T", userInterface)))
					break
				}
				usersWithPasswd = append(usersWithPasswd, userML.UserWithPassword(*user))
			}
			if len(usersWithPasswd) > 0 {
				data = usersWithPasswd
			}
		} else {
			zap.L().Error("error on converting the data to user slice, continue with default data type", zap.String("inputType", fmt.Sprintf("%T", data)))
		}
	}

	var dataBytes []byte
	var err error
	switch provider {
	case backupML.TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", provider), zap.Error(err))
			return
		}
	case backupML.TypeYAML:
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", provider), zap.Error(err))
			return
		}

	default:
		zap.L().Error("This format not supported", zap.String("format", provider), zap.Error(err))
		return
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, backupML.EntityNameIndexSplit, index, provider)
	dir := s.getStorageLocation(provider)
	err = utils.WriteFile(dir, filename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}

func (s *Store) getStorageLocation(provider string) string {
	return fmt.Sprintf("%s/%s/%s", model.GetDirectoryStorage(), s.Config.DumpDir, provider)
}
