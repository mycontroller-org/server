package task

import (
	"sync"

	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type Store struct {
	tasks        map[string]taskTY.Config
	pollingTasks []string // tasks which is in polling mode (will not trigger on events)
	mutex        sync.Mutex
	logger       *zap.Logger
	bus          busTY.Plugin
}

// Add a task
func (s *Store) Add(task taskTY.Config, schedule func(scheduleType, interval string, task *taskTY.Config) string) {
	if !task.Enabled {
		if task.ReEnable { // schedule re-enable job
			schedule(scheduleTypeReEnable, task.ReEnableDuration, &task)
		}
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if task.TriggerOnEvent {
		s.tasks[task.ID] = task
	} else {
		scheduleID := schedule(scheduleTypePolling, task.ExecutionInterval, &task)
		if !utils.ContainsString(s.pollingTasks, scheduleID) {
			s.pollingTasks = append(s.pollingTasks, scheduleID)
		}
	}
}

// UpdateState of a task
func (s *Store) UpdateState(id string, state *taskTY.State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.State = state
	}
	busUtils.SetTaskState(s.logger, s.bus, id, *state)
}

// Remove a task
func (s *Store) Remove(taskID string, unscheduleAll func(scheduleID string), getScheduleId func(IDs ...string) string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	scheduleID := getScheduleId(schedulePrefix, taskID, scheduleTypePolling)
	unscheduleAll(taskID)
	if utils.ContainsString(s.pollingTasks, scheduleID) {
		updatedSlice := make([]string, 0)
		updatedSlice = append(updatedSlice, s.pollingTasks...)
		s.pollingTasks = updatedSlice
	}
	delete(s.tasks, taskID)
}

// GetByID returns handler by id
func (s *Store) Get(id string) taskTY.Config {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.tasks[id]
}

func (s *Store) RemoveAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tasks = make(map[string]taskTY.Config)
}

func (s *Store) ListIDs() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ids := make([]string, 0)
	for id := range s.tasks {
		ids = append(ids, id)
	}
	return ids
}

func (s *Store) filterTasks(evnWrapper *eventWrapper) []taskTY.Config {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filteredTasks := make([]taskTY.Config, 0)
	for id := range s.tasks {
		task := s.tasks[id]

		// if the task is not event based, do not include
		if !task.TriggerOnEvent {
			continue
		}

		// if event filter added and matching do not include
		eventTypes := task.EventFilter.EventTypes
		if len(eventTypes) > 0 && !utils.ContainsString(eventTypes, evnWrapper.Event.Type) {
			continue
		}
		entityTypes := task.EventFilter.EntityTypes
		if len(entityTypes) > 0 && !utils.ContainsString(entityTypes, evnWrapper.Event.EntityType) {
			continue
		}

		filters := s.getFilters(task.EventFilter.Filters)
		matching := false
		s.logger.Debug("filterTasks", zap.Any("filters", filters), zap.Any("event", evnWrapper.Event))

		if len(filters) == 0 {
			matching = true
		} else {
			s.logger.Debug("filterTasks", zap.Any("filters", filters), zap.Any("event", evnWrapper.Event))
			matching = filterUtils.IsMatching(evnWrapper.Event.Entity, filters)
		}
		if matching {
			filteredTasks = append(filteredTasks, task)
		}
	}
	return filteredTasks
}

func (s *Store) getFilters(filtersMap map[string]string) []storageTY.Filter {
	filters := make([]storageTY.Filter, 0)
	for k, v := range filtersMap {
		filters = append(filters, storageTY.Filter{Key: k, Operator: storageTY.OperatorEqual, Value: v})
	}
	return filters
}
