package task

import (
	"fmt"

	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	scheduleUtils "github.com/mycontroller-org/server/v2/pkg/utils/schedule"
	"go.uber.org/zap"
)

const (
	schedulePrefix             = "mc_task_schedule"
	scheduleTypePolling        = "polling"
	scheduleTypeActiveDuration = "active_duration"
	scheduleTypeReEnable       = "re_enable"
)

func schedule(scheduleType, interval string, task *taskTY.Config) string {
	switch scheduleType {
	case scheduleTypePolling:
		return scheduleTask(task, scheduleType, interval, pollingTaskTriggerFunc(task))

	case scheduleTypeActiveDuration:
		return scheduleTask(task, scheduleType, interval, taskActiveDurationFunc(task))

	case scheduleTypeReEnable:
		return scheduleTask(task, scheduleType, interval, taskReEnableFunc(task))

	default:
		// noop
		return ""
	}
}

func unschedule(scheduleID string) {
	scheduleUtils.Unschedule(scheduleID)
	zap.L().Debug("removed a task from scheduler", zap.String("scheduleId", scheduleID))
}

func unscheduleAll(taskID string) {
	if taskID == "" {
		scheduleUtils.UnscheduleAll(schedulePrefix)
	} else {
		scheduleUtils.UnscheduleAll(schedulePrefix, taskID)
	}
}

func scheduleTask(task *taskTY.Config, scheduleType string, interval string, callBackFn func()) string {
	if task.State == nil {
		task.State = &taskTY.State{}
	}

	scheduleID := scheduleUtils.GetScheduleID(schedulePrefix, task.ID, scheduleType)
	cronSpec := fmt.Sprintf("@every %s", interval)
	err := scheduleUtils.Schedule(scheduleID, cronSpec, callBackFn)
	if err != nil {
		zap.L().Error("error on adding a task into scheduler", zap.String("scheduleType", scheduleType), zap.String("id", task.ID), zap.String("executionInterval", task.ExecutionInterval), zap.Error(err))
		task.State.LastStatus = false
		task.State.Message = fmt.Sprintf("Error on adding into scheduler: %s", err.Error())
		busUtils.SetTaskState(task.ID, *task.State)
		return ""
	}
	zap.L().Debug("added a task into schedule", zap.String("taskId", task.ID), zap.String("scheduleType", scheduleType), zap.String("scheduleId", scheduleID), zap.String("id", task.ID), zap.Any("cronSpec", cronSpec))
	task.State.Message = fmt.Sprintf("Added into scheduler. cron spec:[%s], scheduleType:%s", cronSpec, scheduleType)
	busUtils.SetTaskState(task.ID, *task.State)
	return scheduleID
}

// verify task active duration func
func taskActiveDurationFunc(task *taskTY.Config) func() {
	scheduleID := scheduleUtils.GetScheduleID(schedulePrefix, task.ID, scheduleTypeActiveDuration)
	taskID := task.ID
	return func() {
		// remove the schedule
		unschedule(scheduleID)
		// execute active task
		zap.L().Debug("verifying a active duration dampening task", zap.String("id", taskID))
		executeTask(task, nil)
	}
}

func pollingTaskTriggerFunc(task *taskTY.Config) func() {
	return func() {
		zap.L().Debug("executing a task by polling", zap.String("id", task.ID))
		executeTask(task, nil)
	}
}

func taskReEnableFunc(task *taskTY.Config) func() {
	scheduleID := scheduleUtils.GetScheduleID(schedulePrefix, task.ID, scheduleTypeReEnable)
	taskID := task.ID
	return func() {
		// remove the schedule
		unschedule(scheduleID)
		zap.L().Debug("re-enabling a task", zap.String("id", taskID))
		busUtils.EnableTask(taskID)
	}
}
