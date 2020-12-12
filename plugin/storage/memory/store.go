package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	exportml "github.com/mycontroller-org/backend/v2/pkg/model/export"
	"github.com/mycontroller-org/backend/v2/pkg/scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/util"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	defaultSyncInterval = "1m"
	syncJobName         = "in-memory-db-sync-to-disk"

	defaultDumpFormat = exportml.TypeJSON
)

// Config of the memory storage
type Config struct {
	DumpEnabled  bool     `yaml:"dump_enabled"`
	DumpInterval string   `yaml:"dump_interval"`
	DumpDir      string   `yaml:"dump_dir"`
	DumpFormat   []string `yaml:"dump_format"`
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

	if cfg.DumpEnabled {
		if cfg.DumpDir == "" {
			return nil, errors.New("dump_dir not defined, but memory database dump enabled")
		}
		// update default dump format, if none supplied
		if len(cfg.DumpFormat) == 0 {
			cfg.DumpFormat = []string{defaultDumpFormat}
		}
		if err != nil {
			return nil, err
		}
		// add sync job
		if cfg.DumpInterval == "" {
			cfg.DumpInterval = defaultSyncInterval
		}
		err = sch.AddFunc(syncJobName, fmt.Sprintf("@every %s", cfg.DumpInterval), store.writeToDisk)
		if err != nil {
			return nil, err
		}
		zap.L().Debug("Memory database dump job scheduled", zap.Any("config", cfg))
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
				positionStart := index * exportml.LimitPerFile
				index++
				positionEnd := (index * exportml.LimitPerFile)

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
	case exportml.TypeJSON:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", provider), zap.Error(err))
			return
		}
	case exportml.TypeYAML:
		dataBytes, err = yaml.Marshal(data)
		if err != nil {
			zap.L().Error("failed to convert to target format", zap.String("format", provider), zap.Error(err))
			return
		}

	default:
		zap.L().Error("This format not supported", zap.String("format", provider), zap.Error(err))
		return
	}

	filename := fmt.Sprintf("%s%s%d.%s", entityName, exportml.EntityNameIndexSplit, index, provider)
	dir := fmt.Sprintf("%s/%s", s.Config.DumpDir, provider)
	err = util.WriteFile(dir, filename, dataBytes)
	if err != nil {
		zap.L().Error("failed to write data to disk", zap.String("directory", dir), zap.String("filename", filename), zap.Error(err))
	}
}
