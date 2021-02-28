package task

import (
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"go.uber.org/zap"
)

func getTaskPollingTriggerFunc(task *taskML.Config, spec string) func() {
	return func() { taskPollingTriggerFunc(task, spec) }
}

func taskPollingTriggerFunc(task *taskML.Config, spec string) {
	zap.L().Debug("executing as task on polling", zap.String("id", task.ID))
	executeTask(task, nil)
}
