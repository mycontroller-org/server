package systemjobs

import (
	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	"github.com/mycontroller-org/server/v2/pkg/model"
	scheduleML "github.com/mycontroller-org/server/v2/pkg/model/schedule"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	"go.uber.org/zap"
)

// updateSunriseSchedules func
func updateSunriseSchedules() {
	filters := []stgType.Filter{{Key: model.KeyScheduleType, Operator: stgType.OperatorIn, Value: []string{scheduleML.TypeSunrise, scheduleML.TypeSunset}}}
	pagination := &stgType.Pagination{Limit: 100}
	result, err := scheduleAPI.List(filters, pagination)
	if err != nil {
		zap.L().Error("error on fetching schedule jobs", zap.Error(err))
	}
	if result.Count == 0 {
		return
	}
	schedules, ok := result.Data.(*[]scheduleML.Config)
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
}
