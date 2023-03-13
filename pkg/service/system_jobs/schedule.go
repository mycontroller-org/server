package systemjobs

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	systemJobPrefix = "system_job"
)

// Schedule a job
func (svc *SystemJobsService) schedule(id, cronSpec string, callBack func()) {
	if cronSpec == "" {
		return
	}
	svc.unschedule(id)

	name := svc.getScheduleID(id)
	err := svc.scheduler.AddFunc(name, cronSpec, callBack)
	if err != nil {
		svc.logger.Error("error on adding system schedule", zap.Error(err))
		return
	}
	svc.logger.Debug("added a system schedule", zap.String("name", name), zap.String("ID", id), zap.Any("cronSpec", cronSpec))
}

// Unschedule a job
func (svc *SystemJobsService) unschedule(id string) {
	name := svc.getScheduleID(id)
	svc.scheduler.RemoveFunc(name)
	svc.logger.Debug("removed a schedule", zap.String("name", name), zap.String("id", id))
}

// get schedule id
func (svc *SystemJobsService) getScheduleID(id string) string {
	return fmt.Sprintf("%s_%s", systemJobPrefix, id)
}
