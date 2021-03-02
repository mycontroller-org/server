package systemjobs

import (
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// updateSunriseSchedules func
func updateSunriseSchedules() {
	filters := []stgML.Filter{{Key: model.KeyScheduleType, Operator: stgML.OperatorIn, Value: []string{schedulerML.TypeSunrise, schedulerML.TypeSunset}}}
	pagination := &stgML.Pagination{Limit: 100}
	result, err := schedulerAPI.List(filters, pagination)
	if err != nil {
		zap.L().Error("error on fetching schedule jobs", zap.Error(err))
	}
	if result.Count == 0 {
		return
	}
	schedules, ok := result.Data.(*[]schedulerML.Config)
	if !ok {
		zap.L().Error("error on converting to target type")
		return
	}

	scheduleIDs := []string{}
	for index := 0; index < len(*schedules); index++ {
		schedule := (*schedules)[index]
		scheduleIDs = append(scheduleIDs, schedule.ID)
	}

	err = schedulerAPI.Reload(scheduleIDs)
	if err != nil {
		zap.L().Error("error on reloading schedules", zap.Error(err))
	}
}
