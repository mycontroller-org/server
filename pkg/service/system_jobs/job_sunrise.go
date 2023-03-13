package systemjobs

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const (
	idSunrise = "sunrise"
)

// ReloadJob func
func (svc *SystemJobsService) reloadSunriseJob() {
	jobs, err := svc.api.Settings().GetSystemJobs()
	if err != nil {
		svc.logger.Error("error on getting system jobs", zap.Error(err))
	}

	// update sunrise job
	svc.schedule(idSunrise, jobs.Sunrise, svc.updateSunriseSchedules)
}

// updateSunriseSchedules func
func (svc *SystemJobsService) updateSunriseSchedules() {
	filters := []storageTY.Filter{{Key: types.KeyScheduleType, Operator: storageTY.OperatorIn, Value: []string{schedulerTY.TypeSunrise, schedulerTY.TypeSunset}}}
	limit := int64(50)
	offset := int64(0)
	pagination := &storageTY.Pagination{
		Limit:  limit,
		Offset: offset,
		SortBy: []storageTY.Sort{{Field: types.KeyID, OrderBy: storageTY.SortByASC}},
	}
	for {
		pagination.Offset = offset
		result, err := svc.api.Schedule().List(filters, pagination)
		if err != nil {
			svc.logger.Error("error on fetching schedule jobs", zap.Error(err))
		}
		if result.Count == 0 {
			return
		}
		schedules, ok := result.Data.(*[]schedulerTY.Config)
		if !ok {
			svc.logger.Error("error on converting to target type")
			return
		}

		scheduleIDs := []string{}
		for index := 0; index < len(*schedules); index++ {
			schedule := (*schedules)[index]
			scheduleIDs = append(scheduleIDs, schedule.ID)
		}

		err = svc.api.Schedule().Reload(scheduleIDs)
		if err != nil {
			svc.logger.Error("error on reloading schedules", zap.Error(err))
		}

		if result.Count < limit {
			return
		}
		offset++
	}
}
