package task

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	variablesUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

func executeTask(task *taskML.Config, evntWrapper *eventWrapper) {
	start := time.Now()

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
		return
	}

	// user can access event
	if evntWrapper != nil {
		variables[model.KeyEvent] = evntWrapper.Event
	}
	variables[model.KeyTask] = task // include task in to the variables list

	triggered := false
	// execute conditions
	switch task.EvaluationType {
	case taskML.EvaluationTypeRule:
		triggered = isTriggered(task.EvaluationConfig.Rule, variables)

	case taskML.EvaluationTypeJavascript:
		responseMap, triggeredStatus := isTriggeredJavascript(task.ID, task.EvaluationConfig, variables)
		triggered = triggeredStatus
		variables = variablesUtils.Merge(variables, responseMap)

	case taskML.EvaluationTypeWebhook:
		responseMap, triggeredStatus := isTriggeredWebhook(task.ID, task.EvaluationConfig, variables)
		triggered = triggeredStatus
		variables = variablesUtils.Merge(variables, responseMap)

	default:
		zap.L().Error("Unknown rule engine type", zap.String("type", task.EvaluationType))
	}

	notifyHandlers := false

	// if ignoreDuplicate enabled and last status false
	if triggered && task.IgnoreDuplicate && !state.LastStatus {
		notifyHandlers = true
	}

	// if triggered and ignoreDuplicate disabled
	if triggered && !task.IgnoreDuplicate {
		notifyHandlers = true
	}

	if notifyHandlers {
		state.LastSuccess = start // update last success time
		state.Executions = append(state.Executions, true)

		parameters := variablesUtils.UpdateParameters(variables, task.HandlerParameters)
		variablesUtils.UpdateParameters(variables, parameters)
		finalData := variablesUtils.MergeParameter(variables, parameters)
		busUtils.PostToHandler(task.Handlers, finalData)
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
		busUtils.DisableTask(task.ID)
	}
}
