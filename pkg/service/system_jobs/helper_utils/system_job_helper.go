package helper

import (
	"fmt"

	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"

	"go.uber.org/zap"
)

const (
	systemJobPrefix = "system_job"
)

// Schedule a job
func Schedule(id, cronSpec string, callBack func()) {
	if cronSpec == "" {
		return
	}
	Unschedule(id)

	name := GetScheduleID(id)
	err := coreScheduler.SVC.AddFunc(name, cronSpec, callBack)
	if err != nil {
		zap.L().Error("error on adding system schedule", zap.Error(err))
		return
	}
	zap.L().Debug("added a system schedule", zap.String("name", name), zap.String("ID", id), zap.Any("cronSpec", cronSpec))
}

// Unschedule a job
func Unschedule(id string) {
	name := GetScheduleID(id)
	coreScheduler.SVC.RemoveFunc(name)
	zap.L().Debug("removed a schedule", zap.String("name", name), zap.String("id", id))
}

// get schedule id
func GetScheduleID(id string) string {
	return fmt.Sprintf("%s_%s", systemJobPrefix, id)
}
