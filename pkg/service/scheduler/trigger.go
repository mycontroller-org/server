package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	variablesUtils "github.com/mycontroller-org/server/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

func (svc *SchedulerService) getScheduleTriggerFunc(cfg *schedulerTY.Config, spec string) func() {
	return func() { svc.scheduleTriggerFunc(cfg, spec) }
}

func (svc *SchedulerService) scheduleTriggerFunc(cfg *schedulerTY.Config, spec string) {
	// validate schedule
	if !svc.isValidSchedule(cfg) {
		svc.logger.Debug("at this time, this is not a valid schedule", zap.String("ScheduleID", cfg.ID), zap.String("spec", spec), zap.Any("validity details", cfg.Validity))
		return
	}

	start := time.Now()

	cfg.State.LastRun = time.Now()
	cfg.State.ExecutedCount++
	cfg.State.LastStatus = true
	cfg.State.Message = ""
	svc.logger.Debug("triggered", zap.String("ID", cfg.ID), zap.String("spec", spec))

	executionError := ""

	// disable even there is a error on the schedule
	// call it inside func to get updated "executionError" value
	defer func() { svc.verifyAndDisableSchedule(cfg, time.Since(start), executionError) }()

	// load variables
	variables, err := svc.variablesEngine.Load(cfg.Variables)
	if err != nil {
		svc.logger.Error("error on loading variables", zap.String("schedulerID", cfg.ID), zap.Error(err))
		// update triggered count and update state
		cfg.State.LastStatus = false
		cfg.State.Message = fmt.Sprintf("error: %s", err.Error())
		busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
		executionError = err.Error()
		return
	}

	variables[types.KeySchedule] = cfg // include schedule in to the variables list

	switch cfg.CustomVariableType {
	case schedulerTY.CustomVariableTypeNone, "":
		// no action needed

	case schedulerTY.CustomVariableTypeJavascript:
		if cfg.CustomVariableConfig.Javascript != "" {

			// add script timeout from label
			var scriptTimeout *time.Duration
			if cfg.Labels.IsExists(types.LabelScriptTimeout) && cfg.Labels.Get(types.LabelScriptTimeout) != "" {
				timeoutStr := cfg.Labels.Get(types.LabelScriptTimeout)
				timeoutDuration, err := time.ParseDuration(timeoutStr)
				if err != nil {
					svc.logger.Error("error on parsing script timeout", zap.String("schedule", cfg.ID), zap.String("label", types.LabelScriptTimeout), zap.Error(err))
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

			result, err := javascript.Execute(svc.logger, cfg.CustomVariableConfig.Javascript, variables, scriptTimeout)
			if err != nil {
				svc.logger.Error("error on executing javascript", zap.String("schedulerID", cfg.ID), zap.Error(err))
				cfg.State.LastStatus = false
				cfg.State.Message = fmt.Sprintf("error: %s", err.Error())
				busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
				executionError = err.Error()
				return
			}

			// if the response is a map type merge it with variables
			if resultMap, ok := result.(map[string]interface{}); ok {
				variables = variablesUtils.Merge(variables, resultMap)
			}
		}

	case schedulerTY.CustomVariableTypeWebhook:
		customMap := svc.loadWebhookVariables(cfg.ID, cfg.CustomVariableConfig, variables)
		if len(customMap) > 0 {
			variables = variablesUtils.Merge(variables, customMap)
		}

	default:
		svc.logger.Error("unknown custom variable loader type", zap.String("type", cfg.CustomVariableType))
	}

	// post to handlers
	parameters := variablesUtils.UpdateParameters(svc.logger, variables, cfg.HandlerParameters, svc.templateEngine)
	busUtils.PostToHandler(svc.logger, svc.bus, cfg.Handlers, parameters)

	cfg.State.Message = fmt.Sprintf("time taken: %s", time.Since(start).String())
	// update triggered count and update state
	busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
}

func (svc *SchedulerService) verifyAndDisableSchedule(cfg *schedulerTY.Config, timeTaken time.Duration, executionError string) {
	switch cfg.Type {

	// if repeat type job, verify repeat count
	case schedulerTY.TypeRepeat:
		spec := &schedulerTY.SpecRepeat{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			svc.logger.Error("error on convert to repeat spec", zap.Error(err), zap.String("ScheduleID", cfg.ID), zap.Any("spec", cfg.Spec))
			return
		}
		if spec.RepeatCount != 0 && cfg.State.ExecutedCount >= spec.RepeatCount {
			svc.logger.Debug("reached the maximum execution count, disabling this job", zap.String("ScheduleID", cfg.ID), zap.Any("spec", cfg.Spec))
			busUtils.DisableSchedule(svc.logger, svc.bus, cfg.ID)
			// Sometimes setState updates as enabled
			// To avoid this adding small sleep, but this is not good fix.
			utils.SmartSleep(200 * time.Millisecond)
			cfg.State.Message = fmt.Sprintf("time taken: %s, reached maximum execution count", timeTaken.String())
			if executionError != "" {
				cfg.State.Message = fmt.Sprintf("%s, executionError:%s", cfg.State.Message, executionError)
			}
			busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
			return
		}

	// disable the schedule if it is a on date job
	case schedulerTY.TypeSimple, schedulerTY.TypeSunrise, schedulerTY.TypeSunset:
		spec := &schedulerTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			svc.logger.Error("error on loading spec", zap.String("schedulerID", cfg.ID), zap.Error(err))
			cfg.State.LastStatus = false
			cfg.State.Message = fmt.Sprintf("error: %s", err.Error())
			if executionError != "" {
				cfg.State.Message = fmt.Sprintf("%s, executionError:%s", cfg.State.Message, executionError)
			}
			busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
			return
		}
		if spec.Frequency == schedulerTY.FrequencyOnDate {
			busUtils.DisableSchedule(svc.logger, svc.bus, cfg.ID)
		}
	}

}

func (svc *SchedulerService) getCronSpec(cfg *schedulerTY.Config) (string, error) {
	switch cfg.Type {
	case schedulerTY.TypeRepeat:
		spec := &schedulerTY.SpecRepeat{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("@every %s", spec.Interval), nil

	case schedulerTY.TypeCron:
		spec := &schedulerTY.SpecCron{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return spec.CronExpression, nil

	case schedulerTY.TypeSimple:
		spec := &schedulerTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return toCronExpression(spec)

	case schedulerTY.TypeSunrise, schedulerTY.TypeSunset:
		spec := &schedulerTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}

		var suntime *time.Time

		if cfg.Type == schedulerTY.TypeSunrise {
			sunrise, err := svc.sunriseApi.SunriseTime()
			if err != nil {
				return "", err
			}
			suntime = sunrise
		}
		if cfg.Type == schedulerTY.TypeSunset {
			sunset, err := svc.sunriseApi.SunsetTime()
			if err != nil {
				return "", err
			}
			suntime = sunset
		}
		offset, err := time.ParseDuration(spec.Offset)
		if err != nil {
			return "", err
		}

		updatedTime := suntime.Add(offset)
		spec.Time = updatedTime.Format("15:04:05")
		return toCronExpression(spec)

	default:
		return "", fmt.Errorf("invalid schedule type: %s", cfg.Type)
	}
}

func toCronExpression(spec *schedulerTY.SpecSimple) (string, error) {
	cronRaw := struct {
		Seconds    string
		Minutes    string
		Hours      string
		DayOfMonth string
		Month      string
		DayOfWeek  string
	}{}

	cronRaw.Month = "*"

	switch spec.Frequency {
	case schedulerTY.FrequencyDaily, schedulerTY.FrequencyWeekly:
		cronRaw.DayOfMonth = "*"
		cronRaw.DayOfWeek = spec.DayOfWeek

	case schedulerTY.FrequencyMonthly:
		cronRaw.DayOfWeek = "*"
		cronRaw.DayOfMonth = converterUtils.ToString(spec.DateOfMonth)

	case schedulerTY.FrequencyOnDate:
		date, err := time.Parse(dateTimeTY.CustomDateFormat, spec.Date)
		if err != nil {
			return "", err
		}
		cronRaw.DayOfMonth = converterUtils.ToString(date.Day())
		cronRaw.Month = converterUtils.ToString(int(date.Month()))
		cronRaw.DayOfWeek = "*"

	default:
		return "", fmt.Errorf("invalid frequency: %s", spec.Frequency)
	}

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

	// format: "Seconds Minutes Hours DayOfMonth Month DayOfWeek"
	cron := fmt.Sprintf("%s %s %s %s %s %s", cronRaw.Seconds, cronRaw.Minutes, cronRaw.Hours, cronRaw.DayOfMonth, cronRaw.Month, cronRaw.DayOfWeek)
	return cron, nil
}

func (svc *SchedulerService) isValidSchedule(cfg *schedulerTY.Config) bool {
	if !cfg.Validity.Enabled {
		return true
	}

	fromDate := time.Time(cfg.Validity.Date.From.Time)
	toDate := time.Time(cfg.Validity.Date.To.Time)
	fromTime := time.Time(cfg.Validity.Time.From.Time)
	toTime := time.Time(cfg.Validity.Time.To.Time)

	now := time.Now()

	// update from date with time
	if !fromDate.IsZero() {
		if fromTime.IsZero() { // set time to start of the day
			fromTime = time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(),
				0, 0, 0, 0, now.Location())
		} else { // set the time from defined data
			fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(),
				fromTime.Hour(), fromTime.Minute(), fromTime.Second(), fromTime.Nanosecond(),
				now.Location())
		}

		// update timezone to system timezone
		fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(),
			fromDate.Hour(), fromDate.Minute(), fromDate.Second(), fromDate.Nanosecond(),
			now.Location())
	}

	// update to date with time
	if !toDate.IsZero() {
		if toTime.IsZero() { // set the time to end of the day
			toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
				23, 59, 59, 999999999, now.Location())
		} else { // set the time from defined data
			toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
				toTime.Hour(), toTime.Minute(), toTime.Second(), toTime.Nanosecond(),
				now.Location())
		}

		// update timezone to system timezone
		toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
			toDate.Hour(), toDate.Minute(), toDate.Second(), toDate.Nanosecond(),
			now.Location())
	}

	// validate from date and time
	if !fromDate.IsZero() && now.Before(fromDate) {
		svc.logger.Debug("failed", zap.Any("fromDate", fromDate), zap.Any("now", now))
		return false
	}

	// validate to date and time
	if !toDate.IsZero() && now.After(toDate) {
		svc.logger.Debug("failed", zap.Any("toDate", toDate), zap.Any("now", now))
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
				svc.logger.Debug("failed", zap.Any("fromTime", fromTime))
				return false
			}
		}

		// validate to time
		if !toTime.IsZero() {
			toTimeInt, _ := strconv.ParseUint(toTime.Format(timeFormat), 10, 64)
			if nowTimeInt > toTimeInt {
				svc.logger.Debug("failed", zap.Any("toTime", toTime))
				return false
			}
		}
	}
	return true
}

func updateOnDateJobValidity(cfg *schedulerTY.Config) error {
	if cfg.Type == schedulerTY.TypeSimple ||
		cfg.Type == schedulerTY.TypeSunrise ||
		cfg.Type == schedulerTY.TypeSunset {

		spec := &schedulerTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return err
		}
		if spec.Frequency != schedulerTY.FrequencyOnDate {
			return nil
		}
		date, err := time.Parse(dateTimeTY.CustomDateFormat, spec.Date)
		if err != nil {
			return nil
		}
		cfg.Validity.Enabled = true
		cfg.Validity.Date.From = dateTimeTY.CustomDate{Time: date}
		cfg.Validity.Date.To = dateTimeTY.CustomDate{Time: date}
	}
	return nil
}
