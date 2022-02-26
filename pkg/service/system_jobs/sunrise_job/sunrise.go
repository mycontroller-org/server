package systemjobs

import (
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	helper "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/helper_utils"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const (
	idSunrise = "sunrise"
)

// ReloadJob func
func ReloadJob() {
	jobs, err := settingsAPI.GetSystemJobs()
	if err != nil {
		zap.L().Error("error on getting system jobs", zap.Error(err))
	}

	// update sunrise job
	helper.Schedule(idSunrise, jobs.Sunrise, updateSunriseSchedules)
}

// updateSunriseSchedules func
func updateSunriseSchedules() {
	filters := []storageTY.Filter{{Key: types.KeyScheduleType, Operator: storageTY.OperatorIn, Value: []string{scheduleTY.TypeSunrise, scheduleTY.TypeSunset}}}
	limit := int64(50)
	offset := int64(0)
	pagination := &storageTY.Pagination{
		Limit:  limit,
		Offset: offset,
		SortBy: []storageTY.Sort{{Field: types.KeyID, OrderBy: storageTY.SortByASC}},
	}
	for {
		pagination.Offset = offset
		result, err := scheduleAPI.List(filters, pagination)
		if err != nil {
			zap.L().Error("error on fetching schedule jobs", zap.Error(err))
		}
		if result.Count == 0 {
			return
		}
		schedules, ok := result.Data.(*[]scheduleTY.Config)
		if !ok {
			zap.L().Error("error on converting to target type")
			return
		}

		scheduleIDs := []string{}
		for index := 0; index < len(*schedules); index++ {
			schedule := (*schedules)[index]
			scheduleIDs = append(scheduleIDs, schedule.ID)
		}

		err = scheduleAPI.Reload(scheduleIDs)
		if err != nil {
			zap.L().Error("error on reloading schedules", zap.Error(err))
		}

		if result.Count < limit {
			return
		}
		offset++
	}
}
