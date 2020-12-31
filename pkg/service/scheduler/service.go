package scheduler

import (
	"github.com/robfig/cron/v3"
)

// Scheduler service
var (
	SVC *Scheduler
)

// Scheduler struct
type Scheduler struct {
	cron      *cron.Cron
	jobs      map[string]int
	functions map[string]int
}

// Close scheduler
func (s *Scheduler) Close() {
	s.cron.Stop()
}

// Init cron scheduler
func Init() {
	if SVC == nil {
		SVC = &Scheduler{
			cron:      cron.New(cron.WithSeconds()),
			jobs:      map[string]int{},
			functions: map[string]int{},
		}
		SVC.Start()
	}
}

// Start scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
}

// AddFunc function
func (s *Scheduler) AddFunc(name, cron string, targetFunc func()) error {
	id, err := s.cron.AddFunc(cron, targetFunc)
	if err != nil {
		return err
	}
	s.functions[name] = int(id)
	return nil
}

// RemoveFunc function
func (s *Scheduler) RemoveFunc(name string) {
	id, ok := s.functions[name]
	if ok {
		s.cron.Remove(cron.EntryID(id))
	}
}
