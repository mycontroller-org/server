package core_scheduler

import (
	"strings"
	"sync"

	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/robfig/cron/v3"
)

// Scheduler struct
type CoreSchedulerImpl struct {
	cron      *cron.Cron
	functions map[string]int
	mutex     *sync.RWMutex
}

// return new core scheduler
func New() schedulerTY.CoreScheduler {
	return &CoreSchedulerImpl{
		cron:      cron.New(cron.WithSeconds()),
		functions: map[string]int{},
		mutex:     &sync.RWMutex{},
	}
}

func (cs *CoreSchedulerImpl) Name() string {
	return "core_scheduler"
}

// Close scheduler
func (cs *CoreSchedulerImpl) Close() error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.cron.Stop()
	return nil
}

// starts the core scheduler
func (cs *CoreSchedulerImpl) Start() error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.cron.Start()
	return nil
}

// adds a function to scheduler
func (cs *CoreSchedulerImpl) AddFunc(name, spec string, targetFunc func()) error {
	cs.RemoveFunc(name) // remove the existing schedule, in case found

	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	id, err := cs.cron.AddFunc(spec, targetFunc)
	if err != nil {
		return err
	}
	cs.functions[name] = int(id)
	return nil
}

// removes a function from scheduler
func (cs *CoreSchedulerImpl) RemoveFunc(name string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	id, ok := cs.functions[name]
	if ok {
		cs.cron.Remove(cron.EntryID(id))
	}
}

// returns list of available schedule names
func (cs *CoreSchedulerImpl) ListNames() []string {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	names := make([]string, 0)
	for name := range cs.functions {
		names = append(names, name)
	}
	return names
}

// removes all the schedule with the specified prefix
func (cs *CoreSchedulerImpl) RemoveWithPrefix(prefix string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	for name := range cs.functions {
		if strings.HasPrefix(name, prefix) {
			id, ok := cs.functions[name]
			if ok {
				cs.cron.Remove(cron.EntryID(id))
			}
		}
	}
}

// returns the true/false based on the scheduleID availability
func (cs *CoreSchedulerImpl) IsAvailable(scheduleID string) bool {
	return utils.ContainsString(cs.ListNames(), scheduleID)
}
