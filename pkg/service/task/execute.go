package task

import (
	"strings"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	variablesUtils "github.com/mycontroller-org/server/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

const (
	defaultExecutionsSliceLimit = 10
)

func (svc *TaskService) executeTask(task *taskTY.Config, evntWrapper *eventWrapper) {
	start := time.Now()

	state := task.State
	state.ExecutedCount++
	state.LastEvaluation = start

	svc.logger.Debug("executing a task", zap.String("id", task.ID), zap.String("description", task.Description))
	// load variables
	variables, err := svc.variablesEngine.Load(task.Variables)
	if err != nil {
		svc.logger.Warn("failed to load variables", zap.Error(err), zap.String("taskID", task.ID), zap.String("taskDescription", task.Description))
		// update failure message for state and send it
		state.LastStatus = false
		state.Message = "failed to load a variables"
		svc.store.UpdateState(task.ID, state)
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
		triggered = svc.isTriggered(task.EvaluationConfig.Rule, variables)

	case taskTY.EvaluationTypeJavascript:
		// add script timeout from label
		var scriptTimeout *time.Duration
		if task.Labels.IsExists(types.LabelScriptTimeout) && task.Labels.Get(types.LabelScriptTimeout) != "" {
			timeoutStr := task.Labels.Get(types.LabelScriptTimeout)
			timeoutDuration, err := time.ParseDuration(timeoutStr)
			if err != nil {
				svc.logger.Error("error on parsing script timeout", zap.String("task", task.ID), zap.String("label", types.LabelScriptTimeout), zap.Error(err))
			} else {
				scriptTimeout = &timeoutDuration
			}
		}
		// add default timeout, if timeout label not present
		if scriptTimeout == nil {
			defaultTimeout, err := time.ParseDuration(defaultScriptTimeout)
			if err != nil {
				svc.logger.Error("error on parsing default timeout", zap.String("input", defaultScriptTimeout), zap.Error(err))
			} else {
				scriptTimeout = &defaultTimeout
			}
		}
		responseMap, triggeredStatus := svc.isTriggeredJavascript(task.ID, task.EvaluationConfig, variables, scriptTimeout)
		triggered = triggeredStatus
		variables = variablesUtils.Merge(variables, responseMap)

	case taskTY.EvaluationTypeWebhook:
		responseMap, triggeredStatus := svc.isTriggeredWebhook(task.ID, task.EvaluationConfig, variables)
		triggered = triggeredStatus
		variables = variablesUtils.Merge(variables, responseMap)

	default:
		svc.logger.Error("unknown evaluation type", zap.String("type", task.EvaluationType), zap.String("taskId", task.ID))
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
		dampeningTriggered, executionsSliceLimit = svc.executeDampeningConsecutive(task, triggered)

	case taskTY.DampeningTypeEvaluation:
		dampeningTriggered, executionsSliceLimit = svc.executeDampeningEvaluations(task, triggered)

	case taskTY.DampeningTypeActiveDuration:
		dampeningTriggered = svc.executeDampeningActiveDuration(task, triggered)

	default:
		svc.logger.Error("unknown dampening type", zap.String("type", task.Dampening.Type), zap.String("taskId", task.ID))
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
		parameters := variablesUtils.UpdateParameters(svc.logger, variables, task.HandlerParameters, svc.variablesEngine.TemplateEngine())
		variablesUtils.UpdateParameters(svc.logger, variables, parameters, svc.variablesEngine.TemplateEngine())
		busUtils.PostToHandler(svc.logger, svc.bus, task.Handlers, parameters)
	}

	// limit executions status slice
	if len(state.ExecutionsHistory) > executionsSliceLimit {
		state.ExecutionsHistory = state.ExecutionsHistory[:executionsSliceLimit]
	}

	state.LastStatus = dampeningTriggered           // update triggered status
	state.LastDuration = time.Since(start).String() // update last duration
	svc.store.UpdateState(task.ID, state)

	// check autoDisable and re-enable (if applicable)
	if dampeningTriggered && task.AutoDisable {
		busUtils.DisableTask(svc.logger, svc.bus, task.ID)
	}
}

func (svc *TaskService) executeDampeningConsecutive(task *taskTY.Config, triggered bool) (bool, int) {
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

func (svc *TaskService) executeDampeningEvaluations(task *taskTY.Config, triggered bool) (bool, int) {
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
func (svc *TaskService) executeDampeningActiveDuration(task *taskTY.Config, triggered bool) bool {
	scheduleID := svc.getScheduleId(schedulePrefix, task.ID, scheduleTypeActiveDuration)
	if !triggered {
		svc.unschedule(scheduleID)
		task.State.ActiveSince = time.Time{}
		return false
	}

	now := time.Now()
	activeDuration := utils.ToDuration(task.Dampening.ActiveDuration, 0)
	if activeDuration == 0 {
		svc.unschedule(scheduleID)
		svc.logger.Debug("active duration can not be zero in a task", zap.String("id", task.ID), zap.String("activeDuration", task.Dampening.ActiveDuration))
		return false
	}

	activeSince := now.Sub(task.State.ActiveSince)
	// adding 500 millisecond to avoid false on trigger edge
	// Note: in case, if active duration doesn't work properly, revisit activeSince and activeDuration
	activeSince += time.Millisecond * 500
	if activeSince >= activeDuration {
		svc.unschedule(scheduleID)
		return true
	} else {
		if !svc.scheduler.IsAvailable(scheduleID) {
			svc.schedule(scheduleTypeActiveDuration, task.Dampening.ActiveDuration, task)
		}
		return false
	}
}

func (svc *TaskService) getScheduleId(IDs ...string) string {
	return strings.Join(IDs, "_")
}
