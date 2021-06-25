package task

import (
	taskML "github.com/mycontroller-org/server/v2/pkg/model/task"
	"go.uber.org/zap"
)

func getTaskPollingTriggerFunc(task *taskML.Config) func() {
	return func() { taskPollingTriggerFunc(task) }
}

func taskPollingTriggerFunc(task *taskML.Config) {
	zap.L().Debug("executing a task on polling", zap.String("id", task.ID))
	executeTask(task, nil)
}
