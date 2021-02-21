package task

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	variablesUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

func resourcePostProcessor(item interface{}) {
	resource, ok := item.(*resourceWrapper)
	if !ok {
		zap.L().Warn("supplied item is not resourceWrapper", zap.Any("item", item))
		return
	}

	zap.L().Debug("resource received", zap.String("type", resource.ResourceType))

	for index := 0; index < len(resource.Tasks); index++ {
		task := resource.Tasks[index]
		state := task.State
		state.ExecutedCount = state.ExecutedCount + 1
		state.LastEvaluation = time.Now()

		zap.L().Debug("executing a task", zap.String("id", task.ID), zap.String("description", task.Description))
		// load variables
		variables, err := variablesUtils.LoadVariables(task.Variables)
		if err != nil {
			zap.L().Warn("failed to load variables", zap.Error(err), zap.String("taskID", task.ID), zap.String("taskDescription", task.Description))

			// update failure message for state and send it
			state.LastStatus = false
			state.Message = "failed to load a variables"
			tasksStore.UpdateState(task.ID, state)
			continue
		}

		triggered := false
		// execute conditions
		if task.RemoteCall {
			// do remote call things
		} else {
			triggered = isTriggered(task.Rule, variables)
		}

		state.LastStatus = triggered // update triggered status

		// if returns true, execute notifications
		if triggered {
			state.LastSuccess = state.LastEvaluation // update last success time
			zap.L().Debug("Executing notifications", zap.Any("notify", task.Notify))
			variables[model.KeyTask] = task // include task in the variables
			for _, handlerID := range task.Notify {
				postToHandler(handlerID, variables)
			}
		} else {
			zap.L().Debug("Conditions failed", zap.Any("notify", task.Notify))
		}

		state.Message = "" // clear message and send it
		tasksStore.UpdateState(task.ID, state)
	}
}

func postToHandler(handlerID string, variables map[string]interface{}) {
	msg := &handlerML.MessageWrapper{
		ID:        handlerID,
		Variables: variables,
	}
	err := mcbus.Publish(mcbus.FormatTopic(mcbus.TopicPostMessageNotifyHandler), msg)
	if err != nil {
		zap.L().Error("error on message publish", zap.Error(err))
	}
}
