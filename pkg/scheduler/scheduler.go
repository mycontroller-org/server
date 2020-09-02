package scheduler

import (
	"github.com/robfig/cron/v3"
)

// Scheduler struct
type Scheduler struct {
	cron  *cron.Cron
	jobs  map[string]int
	funcs map[string]int
}

// Close scheduler
func (s *Scheduler) Close() {
	s.cron.Stop()
}

// Init cron scheduler
func Init() *Scheduler {
	return &Scheduler{
		cron:  cron.New(cron.WithSeconds()),
		jobs:  map[string]int{},
		funcs: map[string]int{},
	}
}

// Start scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
}

// AddFunc function
func (s *Scheduler) AddFunc(name, cron string, fn func()) error {
	id, err := s.cron.AddFunc(cron, fn)
	if err != nil {
		return err
	}
	s.funcs[name] = int(id)
	return nil
}

// RemoveFunc function
func (s *Scheduler) RemoveFunc(name string) {
	id, ok := s.funcs[name]
	if ok {
		s.cron.Remove(cron.EntryID(id))
	}
}