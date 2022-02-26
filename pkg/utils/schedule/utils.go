package utils

import (
	"strings"

	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"go.uber.org/zap"
)

// Unschedule all the jobs with the following prefix
func UnscheduleAll(prefixSlice ...string) {
	prefix := strings.Join(prefixSlice, "_")
	coreScheduler.SVC.RemoveWithPrefix(prefix)
}

// unschedule a job
func Unschedule(scheduleID string) {
	coreScheduler.SVC.RemoveFunc(scheduleID)
}

// Schedule a job with job spec
func Schedule(scheduleID, jobSpec string, triggerFunc func()) error {
	err := coreScheduler.SVC.AddFunc(scheduleID, jobSpec, triggerFunc)
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
		return err
	}
	zap.L().Debug("added a schedule", zap.String("scheduleID", scheduleID), zap.String("jobSpec", jobSpec))
	return nil
}

// Joins all the ids and returns a string
func GetScheduleID(IDs ...string) string {
	return strings.Join(IDs, "_")
}
