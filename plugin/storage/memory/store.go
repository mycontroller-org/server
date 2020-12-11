package memory

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/util"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	defaultSyncInterval = "1m"
	syncJobName         = "memory_db_sync"

	defaultDumpFormat = "json"
)

// Config of the memory storage
type Config struct {
	DiskLocation string   `yaml:"disk_location"`
	DumpFormat   []string `yaml:"dump_format"`
	SyncInterval string   `yaml:"sync_interval"`
}

// Store to keep all the entities
type Store struct {
	sync.RWMutex
	Config   Config
	data     map[string][]interface{} // entities map with entity name
	lastSync time.Time
}

// NewClient in-memory database
func NewClient(config map[string]interface{}, sch *scheduler.Scheduler) (*Store, error) {
	cfg := Config{}
	err := ut.MapToStruct(ut.TagNameYaml, config, &cfg)
	if err != nil {
		return nil, err
	}

	store := &Store{
		Config:   cfg,
		data:     make(map[string][]interface{}),
		lastSync: time.Now(),
	}

	if cfg.DiskLocation != "" {
		// update default dump format, if none supplied
		if len(cfg.DumpFormat) == 0 {
			cfg.DumpFormat = []string{defaultDumpFormat}
		}
		// load data from disk
		err = store.loadFromDisk()
		if err != nil {
			return nil, err
		}
		// add sync job
		if cfg.SyncInterval == "" {
			cfg.SyncInterval = defaultSyncInterval
		}
		err = sch.AddFunc(syncJobName, fmt.Sprintf("@every %s", cfg.SyncInterval), store.writeToDisk)
		if err != nil {
			return nil, err
		}
	}

	return store, nil
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
				positionStart := index * 50
				index++
				positionEnd := (index * 50)

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
	var dataBytes []byte
	var err error
	switch provider {
	case "json":
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", provider), zap.Error(err))
			return
		}
	case "yaml":
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", provider), zap.Error(err))
			return
		}
	default:
		zap.L().Error("This format not supported", zap.String("format", provider), zap.Error(err))
		return
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, "__", index, provider)
	dir := fmt.Sprintf("%s/%s", s.Config.DiskLocation, provider)
	err = util.WriteFile(dir, filename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}

func (s *Store) loadFromDisk() error {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	zap.L().Debug("Loading data from disk", zap.String("location", s.Config.DiskLocation))
	dirs, err := util.ListDirs(s.Config.DiskLocation)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		switch dir.Name {
		case "json", "yaml":
			//targetDir := fmt.Sprintf("%s/%s", s.Config.DiskLocation, dir.Name)
			//return export.ImportEntities(targetDir, dir.Name)
		default:
			continue
		}
	}
	return nil
}
