package schedule

import (
	"fmt"

	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"go.uber.org/zap"
)

const (
	schedulePrefix = "user_schedule"
)

func schedule(cfg *scheduleTY.Config) {
	if cfg.State == nil {
		cfg.State = &scheduleTY.State{}
	}
	name := getScheduleID(cfg.ID)

	// update validity, if it is on date job, add date(with year) on validity
	err := updateOnDateJobValidity(cfg)
	if err != nil {
		zap.L().Error("error on updating on date job config", zap.Error(err), zap.String("id", cfg.ID))
		return
	}

	cronSpec, err := getCronSpec(cfg)
	if err != nil {
		zap.L().Error("error on creating cron spec", zap.Error(err))
		cfg.State.LastStatus = false
		cfg.State.Message = fmt.Sprintf("Error on cron spec creation: %s", err.Error())
		busUtils.SetScheduleState(cfg.ID, *cfg.State)
		return
	}
	err = coreScheduler.SVC.AddFunc(name, cronSpec, getScheduleTriggerFunc(cfg, cronSpec))
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
		cfg.State.LastStatus = false
		cfg.State.Message = fmt.Sprintf("Error on adding into scheduler: %s", err.Error())
		busUtils.SetScheduleState(cfg.ID, *cfg.State)
	}
	zap.L().Debug("added a schedule", zap.String("name", name), zap.String("ID", cfg.ID), zap.Any("cronSpec", cronSpec))
	cfg.State.Message = fmt.Sprintf("Added into scheduler. cron spec:[%s]", cronSpec)
	busUtils.SetScheduleState(cfg.ID, *cfg.State)
}

func unschedule(id string) {
	name := getScheduleID(id)
	coreScheduler.SVC.RemoveFunc(name)
	zap.L().Debug("removed a schedule", zap.String("name", name), zap.String("id", id))
}

func unloadAll() {
	coreScheduler.SVC.RemoveWithPrefix(schedulePrefix)
}

func getScheduleID(id string) string {
	return fmt.Sprintf("%s_%s", schedulePrefix, id)
}
