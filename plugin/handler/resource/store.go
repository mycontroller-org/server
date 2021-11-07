package resource

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"sync"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	store_filename = "handler/resource"
)

type store struct {
	handlerID string
	jobs      map[string]JobsConfig
	mutex     sync.RWMutex
}

// JobsConfig used to keep the pre-delayed jobs
type JobsConfig struct {
	Name      string                   `yaml:"name"`
	Data      handlerType.ResourceData `yaml:"data"`
	Delay     time.Duration            `yaml:"delay"`
	CreatedAt time.Time                `yaml:"created_at"`
}

func (s *store) getName() string {
	return fmt.Sprintf("%s/%s_%s.yaml", model.GetDirectoryDataRoot(), store_filename, s.handlerID)
}

func (s *store) remove(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.jobs, name)
}

func (s *store) add(name string, rsData handlerType.ResourceData, delayStr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		zap.L().Error("invalid delay", zap.String("delayString", delayStr))
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
		return nil
	}

	data, err := ioutil.ReadFile(s.getName())
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

	if len(s.jobs) == 0 {
		return nil
	}

	data, err := yaml.Marshal(s.jobs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(s.getName(), data, fs.ModePerm)
}
