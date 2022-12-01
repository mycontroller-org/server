package task

import (
	"time"

	commonStore "github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	scheduleUtils "github.com/mycontroller-org/server/v2/pkg/utils/schedule"
	variablesUtils "github.com/mycontroller-org/server/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

const (
	defaultExecutionsSliceLimit = 10
)

func executeTask(task *taskTY.Config, evntWrapper *eventWrapper) {
	start := time.Now()

	state := task.State
	state.ExecutedCount++
	state.LastEvaluation = start

	zap.L().Debug("executing a task", zap.String("id", task.ID), zap.String("description", task.Description))
	// load variables
	variables, err := variablesUtils.LoadVariables(task.Variables, commonStore.CFG.Secret)
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
		variables[types.KeyTaskEvent] = evntWrapper.Event
	}
	variables[types.KeyTask] = task // include task in to the variables list

	triggered := false
	// execute conditions
	switch task.EvaluationType {
	case taskTY.EvaluationTypeRule:
		triggered = isTriggered(task.EvaluationConfig.Rule, variables)

	case taskTY.EvaluationTypeJavascript:
		responseMap, triggeredStatus := isTriggeredJavascript(task.ID, task.EvaluationConfig, variables)
		triggered = triggeredStatus
		variables = variablesUtils.Merge(variables, responseMap)

	case taskTY.EvaluationTypeWebhook:
		responseMap, triggeredStatus := isTriggeredWebhook(task.ID, task.EvaluationConfig, variables)
		triggered = triggeredStatus
		variables = variablesUtils.Merge(variables, responseMap)

	default:
		zap.L().Error("unknown evaluation type", zap.String("type", task.EvaluationType), zap.String("taskId", task.ID))
		return
	}

	// update active from
	if !triggered {
		state.ActiveSince = time.Time{}
	} else if state.ActiveSince.IsZero() {
		state.ActiveSince = start
	}

	// update triggered status
	state.ExecutionsHistory = append(state.ExecutionsHistory, taskTY.ExecutionState{Triggered: triggered, Timestamp: start})

	dampeningTriggered := false
	executionsSliceLimit := defaultExecutionsSliceLimit

	switch task.Dampening.Type {
	case taskTY.DampeningTypeNone, "": // none or empty string type
		dampeningTriggered = triggered

	case taskTY.DampeningTypeConsecutive:
		dampeningTriggered, executionsSliceLimit = executeDampeningConsecutive(task, triggered)

	case taskTY.DampeningTypeEvaluation:
		dampeningTriggered, executionsSliceLimit = executeDampeningEvaluations(task, triggered)

	case taskTY.DampeningTypeActiveDuration:
		dampeningTriggered = executeDampeningActiveDuration(task, triggered)

	default:
		zap.L().Error("unknown dampening type", zap.String("type", task.Dampening.Type), zap.String("taskId", task.ID))
		return
	}

	notifyHandlers := false

	if dampeningTriggered {
		if !task.IgnoreDuplicate { // if ignoreDuplicate disabled
			notifyHandlers = true
		} else if !state.LastStatus { // if ignoreDuplicate and last status false
			notifyHandlers = true
		}
	}

	if notifyHandlers {
		state.LastSuccess = start // update last success time
		parameters := variablesUtils.UpdateParameters(variables, task.HandlerParameters)
		variablesUtils.UpdateParameters(variables, parameters)
		busUtils.PostToHandler(task.Handlers, parameters)
	}

	// limit executions status slice
	if len(state.ExecutionsHistory) > executionsSliceLimit {
		state.ExecutionsHistory = state.ExecutionsHistory[:executionsSliceLimit]
	}

	state.LastStatus = dampeningTriggered           // update triggered status
	state.LastDuration = time.Since(start).String() // update last duration
	tasksStore.UpdateState(task.ID, state)

	// check autoDisable and re-enable (if applicable)
	if dampeningTriggered && task.AutoDisable {
		busUtils.DisableTask(task.ID)
	}
}

func executeDampeningConsecutive(task *taskTY.Config, triggered bool) (bool, int) {
	occurrences := int(task.Dampening.Occurrences)
	if !triggered || len(task.State.ExecutionsHistory) < occurrences {
		return false, occurrences
	}
	results := task.State.ExecutionsHistory[:occurrences]
	for _, r := range results {
		if !r.Triggered {
			return false, occurrences
		}
	}
	return true, occurrences // reset executions slice
}

func executeDampeningEvaluations(task *taskTY.Config, triggered bool) (bool, int) {
	occurrences := int(task.Dampening.Occurrences)
	evaluations := int(task.Dampening.Evaluation)
	if !triggered || len(task.State.ExecutionsHistory) < occurrences {
		return false, evaluations
	}
	results := task.State.ExecutionsHistory
	if len(task.State.ExecutionsHistory) >= evaluations {
		results = task.State.ExecutionsHistory[:evaluations]
	}

	occurrenceCount := 0
	for _, r := range results {
		if r.Triggered {
			occurrenceCount++
		}
		if occurrenceCount >= occurrences {
			return true, evaluations // reset executions slice
		}
	}
	return false, evaluations
}

// verifies the active duration dampening
func executeDampeningActiveDuration(task *taskTY.Config, triggered bool) bool {
	scheduleID := scheduleUtils.GetScheduleID(schedulePrefix, task.ID, scheduleTypeActiveDuration)
	if !triggered {
		unschedule(scheduleID)
		task.State.ActiveSince = time.Time{}
		return false
	}

	now := time.Now()
	activeDuration := utils.ToDuration(task.Dampening.ActiveDuration, 0)
	if activeDuration == 0 {
		unschedule(scheduleID)
		zap.L().Debug("active duration can not be zero in a task", zap.String("id", task.ID), zap.String("activeDuration", task.Dampening.ActiveDuration))
		return false
	}

	activeSince := now.Sub(task.State.ActiveSince)
	// adding 500 millisecond to avoid false on trigger edge
	// Note: in case, if active duration doesn't work properly, revisit activeSince and activeDuration
	activeSince += time.Millisecond * 500
	if activeSince >= activeDuration {
		unschedule(scheduleID)
		return true
	} else {
		if !scheduleUtils.IsScheduleAvailable(scheduleID) {
			schedule(scheduleTypeActiveDuration, task.Dampening.ActiveDuration, task)
		}
		return false
	}
}
