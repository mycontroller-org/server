package scheduler

import (
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
)

// Scheduler service
var (
	SVC *Scheduler
)

// Scheduler struct
type Scheduler struct {
	cron      *cron.Cron
	functions map[string]int
	mutex     *sync.RWMutex
}

// Close scheduler
func (s *Scheduler) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cron.Stop()
}

// Init cron scheduler
func Init() {
	if SVC == nil {
		SVC = &Scheduler{
			cron:      cron.New(cron.WithSeconds()),
			functions: map[string]int{},
			mutex:     &sync.RWMutex{},
		}
		SVC.Start()
	}
}

// Start scheduler
func (s *Scheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cron.Start()
}

// AddFunc function
func (s *Scheduler) AddFunc(name, spec string, targetFunc func()) error {
	s.RemoveFunc(name) // remove the existing schedule, in case found

	s.mutex.Lock()
	defer s.mutex.Unlock()
	id, err := s.cron.AddFunc(spec, targetFunc)
	if err != nil {
		return err
	}
	s.functions[name] = int(id)
	return nil
}

// RemoveFunc function
func (s *Scheduler) RemoveFunc(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id, ok := s.functions[name]
	if ok {
		s.cron.Remove(cron.EntryID(id))
	}
}

// GetNames return list of available schedule names
func (s *Scheduler) GetNames() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	names := make([]string, 0)
	for name := range s.functions {
		names = append(names, name)
	}
	return names
}

// RemoveWithPrefix removes all the schedule with the specified prefix
func (s *Scheduler) RemoveWithPrefix(prefix string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for name := range s.functions {
		if strings.HasPrefix(name, prefix) {
			id, ok := s.functions[name]
			if ok {
				s.cron.Remove(cron.EntryID(id))
			}
		}
	}
}
