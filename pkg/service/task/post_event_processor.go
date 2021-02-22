package task

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	rsUtils "github.com/mycontroller-org/backend/v2/pkg/utils/resource_service"
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
		start := time.Now()

		task := resource.Tasks[index]
		state := task.State
		state.ExecutedCount++
		state.LastEvaluation = start

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

		variables[model.KeyTask] = task // include task in to the variables list

		triggered := false
		// execute conditions
		if task.RemoteCall {
			// do remote call things
		} else {
			triggered = isTriggered(task.Rule, variables)
		}

		executeNotify := false

		// if ignoreDuplicate enabled and last status false
		if triggered && task.IgnoreDuplicate && !state.LastStatus {
			executeNotify = true
		}

		// if triggered and ignoreDuplicate disabled
		if triggered && !task.IgnoreDuplicate {
			executeNotify = true
		}

		if executeNotify {
			state.LastSuccess = start // update last success time
			state.Executions = append(state.Executions, true)
			zap.L().Debug("Executing handlers", zap.Any("notify", task.Notify))
			for _, handlerID := range task.Notify {
				postToHandler(handlerID, variables)
			}
		}

		if !triggered {
			state.Executions = append(state.Executions, false)
		}

		// limit executions status array into 10
		if len(state.Executions) > 10 {
			state.Executions = state.Executions[:10]
		}

		state.LastStatus = triggered // update triggered status
		state.Message = fmt.Sprintf("last evaluation timeTaken: %s", time.Since(start).String())
		tasksStore.UpdateState(task.ID, state)

		// check autoDisable
		if triggered && task.AutoDisable {
			rsUtils.DisableTask(task.ID)
		}
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
