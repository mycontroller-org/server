package task

import (
	"fmt"

	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"go.uber.org/zap"
)

const (
	schedulePrefix             = "mc_task_schedule"
	scheduleTypePolling        = "polling"
	scheduleTypeActiveDuration = "active_duration"
	scheduleTypeReEnable       = "re_enable"
)

func (svc *TaskService) schedule(scheduleType, interval string, task *taskTY.Config) string {
	switch scheduleType {
	case scheduleTypePolling:
		return svc.scheduleTask(task, scheduleType, interval, svc.pollingTaskTriggerFunc(task))

	case scheduleTypeActiveDuration:
		return svc.scheduleTask(task, scheduleType, interval, svc.taskActiveDurationFunc(task))

	case scheduleTypeReEnable:
		return svc.scheduleTask(task, scheduleType, interval, svc.taskReEnableFunc(task))

	default:
		// noop
		return ""
	}
}

func (svc *TaskService) unschedule(scheduleID string) {
	svc.scheduler.RemoveFunc(scheduleID)
	svc.logger.Debug("removed a task from scheduler", zap.String("scheduleId", scheduleID))
}

func (svc *TaskService) unscheduleAll(taskID string) {
	if taskID == "" {
		svc.scheduler.RemoveWithPrefix(schedulePrefix)
	} else {
		svc.scheduler.RemoveWithPrefix(svc.getScheduleId(schedulePrefix, taskID))
	}
}

func (svc *TaskService) scheduleTask(task *taskTY.Config, scheduleType string, interval string, callBackFn func()) string {
	if task.State == nil {
		task.State = &taskTY.State{}
	}

	scheduleID := svc.getScheduleId(schedulePrefix, task.ID, scheduleType)
	cronSpec := fmt.Sprintf("@every %s", interval)
	err := svc.scheduler.AddFunc(scheduleID, cronSpec, callBackFn)
	if err != nil {
		svc.logger.Error("error on adding a task into scheduler", zap.String("scheduleType", scheduleType), zap.String("id", task.ID), zap.String("executionInterval", task.ExecutionInterval), zap.Error(err))
		task.State.LastStatus = false
		task.State.Message = fmt.Sprintf("Error on adding into scheduler: %s", err.Error())
		busUtils.SetTaskState(svc.logger, svc.bus, task.ID, *task.State)
		return ""
	}
	svc.logger.Debug("added a task into schedule", zap.String("taskId", task.ID), zap.String("scheduleType", scheduleType), zap.String("scheduleId", scheduleID), zap.String("id", task.ID), zap.Any("cronSpec", cronSpec))
	task.State.Message = fmt.Sprintf("Added into scheduler. cron spec:[%s], scheduleType:%s", cronSpec, scheduleType)
	busUtils.SetTaskState(svc.logger, svc.bus, task.ID, *task.State)
	return scheduleID
}

// verify task active duration func
func (svc *TaskService) taskActiveDurationFunc(task *taskTY.Config) func() {
	scheduleID := svc.getScheduleId(schedulePrefix, task.ID, scheduleTypeActiveDuration)
	taskID := task.ID
	return func() {
		// remove the schedule
		svc.unschedule(scheduleID)
		// execute active task
		svc.logger.Debug("verifying a active duration dampening task", zap.String("id", taskID))
		svc.executeTask(task, nil)
	}
}

func (svc *TaskService) pollingTaskTriggerFunc(task *taskTY.Config) func() {
	return func() {
		svc.logger.Debug("executing a task by polling", zap.String("id", task.ID))
		svc.executeTask(task, nil)
	}
}

func (svc *TaskService) taskReEnableFunc(task *taskTY.Config) func() {
	scheduleID := svc.getScheduleId(schedulePrefix, task.ID, scheduleTypeReEnable)
	taskID := task.ID
	return func() {
		// remove the schedule
		svc.unschedule(scheduleID)
		svc.logger.Debug("re-enabling a task", zap.String("id", taskID))
		busUtils.EnableTask(svc.logger, svc.bus, taskID)
	}
}
