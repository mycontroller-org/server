package task

import (
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"go.uber.org/zap"
)

func getTaskPollingTriggerFunc(task *taskTY.Config) func() {
	return func() { taskPollingTriggerFunc(task) }
}

func taskPollingTriggerFunc(task *taskTY.Config) {
	zap.L().Debug("executing a task on polling", zap.String("id", task.ID))
	executeTask(task, nil)
}
