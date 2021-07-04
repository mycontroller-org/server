package task

import (
	"fmt"
	"sync"

	taskML "github.com/mycontroller-org/server/v2/pkg/model/task"
	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

const (
	schedulePrefix = "task_polling_schedule"
)

type store struct {
	tasks        map[string]taskML.Config
	pollingTasks []string // tasks which is in polling mode (will not trigger on events)
	mutex        sync.Mutex
}

var tasksStore = store{
	tasks:        make(map[string]taskML.Config),
	pollingTasks: make([]string, 0),
}

// Add a task
func (s *store) Add(task taskML.Config) {
	if !task.Enabled {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if task.TriggerOnEvent {
		s.tasks[task.ID] = task
	} else {
		name := schedule(&task)
		if !utils.ContainsString(s.pollingTasks, name) {
			s.pollingTasks = append(s.pollingTasks, name)
		}
	}
}

// UpdateState of a task
func (s *store) UpdateState(id string, state *taskML.State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if task, ok := s.tasks[id]; ok {
		task.State = state
	}
	busUtils.SetTaskState(id, *state)
}

// Remove a task
func (s *store) Remove(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	name := unschedule(id)
	if utils.ContainsString(s.pollingTasks, name) {
		updatedSlice := make([]string, 0)
		updatedSlice = append(updatedSlice, s.pollingTasks...)
		s.pollingTasks = updatedSlice
	}
	delete(s.tasks, id)
}

// GetByID returns handler by id
func (s *store) Get(id string) taskML.Config {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.tasks[id]
}

func (s *store) RemoveAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	tasksStore.tasks = make(map[string]taskML.Config)
}

func (s *store) ListIDs() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ids := make([]string, 0)
	for id := range s.tasks {
		ids = append(ids, id)
	}
	return ids
}

func (s *store) filterTasks(evnWrapper *eventWrapper) []taskML.Config {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filteredTasks := make([]taskML.Config, 0)
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
		zap.L().Debug("filterTasks", zap.Any("filters", filters), zap.Any("event", evnWrapper.Event))

		if len(filters) == 0 {
			matching = true
		} else {
			zap.L().Debug("filterTasks", zap.Any("filters", filters), zap.Any("event", evnWrapper.Event))
			matching = helper.IsMatching(evnWrapper.Event.Entity, filters)
		}
		if matching {
			filteredTasks = append(filteredTasks, task)
		}
	}
	return filteredTasks
}

func (s *store) getFilters(filtersMap map[string]string) []stgType.Filter {
	filters := make([]stgType.Filter, 0)
	for k, v := range filtersMap {
		filters = append(filters, stgType.Filter{Key: k, Operator: stgType.OperatorEqual, Value: v})
	}
	return filters
}

func getScheduleID(id string) string {
	return fmt.Sprintf("%s_%s", schedulePrefix, id)
}

func unschedule(id string) string {
	name := getScheduleID(id)
	coreScheduler.SVC.RemoveFunc(name)
	zap.L().Debug("removed a task from scheduler", zap.String("name", name), zap.String("id", id))
	return name
}

func schedule(task *taskML.Config) string {
	name := getScheduleID(task.ID)
	cronSpec := fmt.Sprintf("@every %s", task.ExecutionInterval)
	err := coreScheduler.SVC.AddFunc(name, cronSpec, getTaskPollingTriggerFunc(task))
	if err != nil {
		zap.L().Error("error on adding a task into scheduler", zap.Error(err), zap.String("id", task.ID), zap.String("executionInterval", task.ExecutionInterval))
		task.State.LastStatus = false
		task.State.Message = fmt.Sprintf("Error on adding into scheduler: %s", err.Error())
		busUtils.SetTaskState(task.ID, *task.State)
	}
	zap.L().Debug("added a task into schedule", zap.String("name", name), zap.String("ID", task.ID), zap.Any("cronSpec", cronSpec))
	task.State.Message = fmt.Sprintf("Added into scheduler. cron spec:[%s]", cronSpec)
	busUtils.SetTaskState(task.ID, *task.State)
	return name
}
