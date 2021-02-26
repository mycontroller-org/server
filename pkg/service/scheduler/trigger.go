package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	variablesUtils "github.com/mycontroller-org/backend/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

func getScheduleTriggerFunc(cfg *schedulerML.Config, spec string) func() {
	return func() { scheduleTriggerFunc(cfg, spec) }
}

func scheduleTriggerFunc(cfg *schedulerML.Config, spec string) {
	// validate schedule
	if !isValidSchedule(cfg) {
		zap.L().Info("at this moment, this is not a valid schedule", zap.String("ScheduleID", cfg.ID), zap.String("spec", spec), zap.Any("validity details", cfg.Validity))
		return
	}

	start := time.Now()

	// if repeat type job, verify repeat count
	if cfg.Type == schedulerML.TypeRepeat {
		spec := &schedulerML.SpecRepeat{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			zap.L().Error("error on convert to repeat spec", zap.Error(err), zap.String("ScheduleID", cfg.ID), zap.Any("spec", cfg.Spec))
			return
		}
		if spec.RepeatCount != 0 && cfg.State.ExecutedCount >= spec.RepeatCount {
			zap.L().Debug("Reached maximum execution count, disbling this job", zap.String("ScheduleID", cfg.ID), zap.Any("spec", cfg.Spec))
			busUtils.DisableSchedule(cfg.ID)
			cfg.State.Message = "Reached maximum execution count"
			busUtils.SetScheduleState(cfg.ID, *cfg.State)
			return
		}
	}

	cfg.State.LastRun = time.Now()
	cfg.State.ExecutedCount++
	cfg.State.LastStatus = true
	cfg.State.Message = ""
	zap.L().Debug("Triggered", zap.String("ID", cfg.ID), zap.String("spec", spec))

	// load variables
	variables, err := variablesUtils.LoadVariables(cfg.Variables)
	if err != nil {
		zap.L().Error("error on loading variables", zap.String("schedulerID", cfg.ID), zap.Error(err))
		// update triggered count and update state
		cfg.State.LastStatus = false
		cfg.State.Message = fmt.Sprintf("Error: %s", err.Error())
		busUtils.SetScheduleState(cfg.ID, *cfg.State)
		return
	}

	// post to handlers
	parameters := variablesUtils.UpdateParameters(variables, cfg.HandlerParameters)
	finalData := variablesUtils.Merge(variables, parameters)
	busUtils.PostToHandler(cfg.Handlers, finalData)

	cfg.State.Message = fmt.Sprintf("time taken: %s", time.Since(start).String())
	// update triggered count and update state
	busUtils.SetScheduleState(cfg.ID, *cfg.State)

}

func getCronSpec(cfg *schedulerML.Config) (string, error) {
	switch cfg.Type {
	case schedulerML.TypeRepeat:
		spec := &schedulerML.SpecRepeat{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("@every %s", spec.Interval), nil

	case schedulerML.TypeCron:
		spec := &schedulerML.SpecCron{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return spec.CronExpression, nil

	case schedulerML.TypeSimple:
		spec := &schedulerML.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return toCronExpression(cfg.Type, spec)
	}
	return "", fmt.Errorf("invalid schedule type: %s", cfg.Type)
}

func toCronExpression(scheduleType string, spec *schedulerML.SpecSimple) (string, error) {
	cronRaw := struct {
		Seconds    string
		Minutes    string
		Hours      string
		DayOfMonth string
		Month      string
		DayOfWeek  string
	}{}

	switch spec.Frequency {
	case schedulerML.FrequencyDaily, schedulerML.FrequencyWeekly:
		cronRaw.DayOfMonth = "*"
		cronRaw.DayOfWeek = spec.DayOfWeek

	case schedulerML.FrequencyMonthly:
		cronRaw.DayOfWeek = "*"
		cronRaw.DayOfMonth = utils.ToString(spec.DateOfMonth)

	default:
		return "", fmt.Errorf("invalid frequency: %s", spec.Frequency)
	}

	switch scheduleType {
	case schedulerML.TypeSimple:
		time := strings.Split(strings.TrimSpace(spec.Time), ":")
		if len(time) > 3 {
			return "", fmt.Errorf("invalid time: %s", spec.Time)
		}

		// update hour and minute
		cronRaw.Hours = time[0]
		cronRaw.Minutes = time[1]

		// update seconds
		if len(time) == 3 {
			cronRaw.Seconds = time[2]
		} else {
			cronRaw.Seconds = "0"
		}

	case schedulerML.TypeSunrise, schedulerML.TypeSunset,
		schedulerML.TypeMoonrise, schedulerML.TypeMoonset:
		// TODO:...
		return "", errors.New("this type not implemented yet")

	default:
		return "", fmt.Errorf("invalid schedule type:%s", scheduleType)
	}

	// format: "Seconds Minutes Hours DayOfMonth Month DayOfWeek"
	cron := fmt.Sprintf("%s %s %s %s * %s", cronRaw.Seconds, cronRaw.Minutes, cronRaw.Hours, cronRaw.DayOfMonth, cronRaw.DayOfWeek)
	return cron, nil
}

func isValidSchedule(cfg *schedulerML.Config) bool {
	fromDate := time.Time(cfg.Validity.Date.From.Time)
	toDate := time.Time(cfg.Validity.Date.To.Time)
	fromTime := time.Time(cfg.Validity.Time.From.Time)
	toTime := time.Time(cfg.Validity.Time.To.Time)

	now := time.Now()

	// update from date with time
	if !fromDate.IsZero() {
		if fromTime.IsZero() { // set time to start of the day
			fromTime = time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(),
				0, 0, 0, 0, fromTime.Location())
		} else { // set the time from defined data
			fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(),
				fromTime.Hour(), fromTime.Minute(), fromTime.Second(), fromTime.Nanosecond(),
				fromDate.Location())
		}
	}

	// update to date with time
	if !toDate.IsZero() {
		if toTime.IsZero() { // set the time to end of the day
			toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
				23, 59, 59, 999999999, toDate.Location())
		} else { // set the time from defined data
			toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
				toTime.Hour(), toTime.Minute(), toTime.Second(), toTime.Nanosecond(),
				toDate.Location())
		}
	}

	// validate from date and time
	if !fromDate.IsZero() && fromDate.After(now) {
		return false
	}

	// validate to date and time
	if !toDate.IsZero() && toDate.Before(now) {
		return false
	}

	// if every date time validation enabled
	if cfg.Validity.ValidateTimeEveryday && (!fromTime.IsZero() || !toTime.IsZero()) {
		timeFormat := "150405"
		nowTimeInt, _ := strconv.ParseUint(now.Format(timeFormat), 10, 64)

		// validate from time
		if !fromTime.IsZero() {
			fromTimeInt, _ := strconv.ParseUint(fromTime.Format(timeFormat), 10, 64)
			if nowTimeInt < fromTimeInt {
				return false
			}
		}

		// validate to time
		if !fromTime.IsZero() {
			toTimeInt, _ := strconv.ParseUint(toTime.Format(timeFormat), 10, 64)
			if nowTimeInt > toTimeInt {
				return false
			}
		}
	}
	return true
}

func postToHandler(handlerID string, variables map[string]interface{}) {
	msg := &handlerML.MessageWrapper{
		ID:   handlerID,
		Data: variables,
	}
	err := mcbus.Publish(mcbus.FormatTopic(mcbus.TopicPostMessageNotifyHandler), msg)
	if err != nil {
		zap.L().Error("error on message publish", zap.Error(err))
	}
}
