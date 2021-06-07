package systemjobs

import (
	"fmt"

	analyticsAPI "github.com/mycontroller-org/backend/v2/pkg/analytics"
	settingsAPI "github.com/mycontroller-org/backend/v2/pkg/api/settings"
	configSVC "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	coreScheduler "github.com/mycontroller-org/backend/v2/pkg/service/core_scheduler"

	"go.uber.org/zap"
)

const (
	systemJobPrefix = "system_job"
	idSunrise       = "sunrise"
	idAnalyticsJob  = "analytics_job"
)

// ReloadSystemJobs func
func ReloadSystemJobs() {
	jobs, err := settingsAPI.GetSystemJobs()
	if err != nil {
		zap.L().Error("error on getting system jobs", zap.Error(err))
	}

	// update sunrise job
	schedule(idSunrise, jobs.Sunrise, updateSunriseSchedules)

	// update analytics job
	unschedule(idAnalyticsJob)
	if configSVC.CFG.Analytics.Enabled {
		zap.L().Info(`
#=========================================================================================================#
#                                   Non-PII data Processing Agreement                                     #
#                                   ---------------------------------                                     #
# Analytics enabled in your server, says you agreed that the non-PII data will be collected, processed,   #
# and used by MyController.org to improve the quality of the software.                                    #
# (non-PII - non Personally Identifiable Information)                                                     #
#                                                                                                         #
# If you do not like to share the non-PII data, you can disable it on the configuration file              #
# and restart the server                                                                                  #
#                                                                                                         #
# file: mycontroller.yaml                                                                                 #
# analytics:                                                                                              #
#   enabled: false                                                                                        #
#=========================================================================================================#
`)
		schedule(idAnalyticsJob, "0 15 1 * * *", analyticsAPI.ReportAnalyticsData) // everyday at 1:15 AM
	}

}

func schedule(id, cronSpec string, callBack func()) {
	if cronSpec == "" {
		return
	}
	unschedule(id)

	name := getScheduleID(id)
	err := coreScheduler.SVC.AddFunc(name, cronSpec, callBack)
	if err != nil {
		zap.L().Error("error on adding system schedule", zap.Error(err))
		return
	}
	zap.L().Debug("added a system schedule", zap.String("name", name), zap.String("ID", id), zap.Any("cronSpec", cronSpec))
}

func unschedule(id string) {
	name := getScheduleID(id)
	coreScheduler.SVC.RemoveFunc(name)
	zap.L().Debug("removed a schedule", zap.String("name", name), zap.String("id", id))
}

func getScheduleID(id string) string {
	return fmt.Sprintf("%s_%s", systemJobPrefix, id)
}
