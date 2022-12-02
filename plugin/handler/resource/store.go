package resource

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	store_directory = "handler"
	store_filename  = "resource"
)

type store struct {
	handlerID string
	jobs      map[string]JobsConfig
	mutex     sync.RWMutex
}

// JobsConfig used to keep the pre-delayed jobs
type JobsConfig struct {
	Name      string                 `json:"name" yaml:"name"`
	Data      handlerTY.ResourceData `json:"date" yaml:"data"`
	Delay     time.Duration          `json:"delay" yaml:"delay"`
	CreatedAt time.Time              `json:"createdAt" yaml:"createdAt"`
}

func (s *store) getName() string {
	dir := filepath.Join(types.GetDirectoryDataRoot(), store_directory)
	err := utils.CreateDir(dir)
	if err != nil {
		zap.L().Error("failed to create handler data persistence directory", zap.String("directory", dir))
	}
	return filepath.Join(dir, fmt.Sprintf("%s_%s.yaml", store_filename, s.handlerID))
}

func (s *store) add(name string, rsData handlerTY.ResourceData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delay, err := time.ParseDuration(rsData.PreDelay)
	if err != nil {
		zap.L().Error("invalid delay", zap.String("quickID", rsData.QuickID), zap.String("preDelay", rsData.PreDelay))
		return
	}

	s.jobs[name] = JobsConfig{
		Name:      name,
		Data:      rsData,
		Delay:     delay,
		CreatedAt: time.Now(),
	}
}

func (s *store) loadFromDisk(client *ResourceClient) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !utils.IsFileExists(s.getName()) {
		zap.L().Debug("File not found", zap.String("filename", s.getName()), zap.String("handler", s.handlerID))
		return nil
	}

	zap.L().Debug("Loading from", zap.String("filename", s.getName()), zap.String("handler", s.handlerID))

	data, err := os.ReadFile(s.getName())
	if err != nil {
		return err
	}
	jobs := map[string]JobsConfig{}
	err = yaml.Unmarshal(data, &jobs)
	if err != nil {
		return err
	}
	s.jobs = jobs

	currentTime := time.Now()
	// load data to scheduler
	for name := range s.jobs {
		job := s.jobs[name]
		jobTime := job.CreatedAt.Add(job.Delay)
		if jobTime.Before(currentTime) { // verify the validity
			zap.L().Debug("Expired job", zap.Any("job", job))
			delete(s.jobs, name)
		} else {
			// update delay time
			newDelay := jobTime.Sub(currentTime)
			job.Data.PreDelay = newDelay.String()
			client.schedule(job.Name, job.Data)
		}
	}

	return nil
}

func (s *store) saveToDisk() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := yaml.Marshal(s.jobs)
	if err != nil {
		return err
	}
	zap.L().Debug("Saving the jobs data", zap.String("filename", s.getName()), zap.String("handler", s.handlerID))
	return os.WriteFile(s.getName(), data, fs.ModePerm)
}
