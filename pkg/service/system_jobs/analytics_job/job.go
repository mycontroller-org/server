package systemjobs

import (
	"fmt"
	"math/rand"

	analyticsAPI "github.com/mycontroller-org/server/v2/pkg/analytics"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	helper "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/helper_utils"
	store "github.com/mycontroller-org/server/v2/pkg/store"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	utils "github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	idAnalyticsJob = "analytics"
)

// ReloadJob func
func ReloadJob() {
	// update analytics job
	analyticsJobName := helper.GetScheduleID(idAnalyticsJob)
	helper.Unschedule(analyticsJobName)
	if store.CFG.Analytics.Enabled {
		// generate analyticsId if not available
		_, err := settingsAPI.GetAnalytics()
		if err != nil {
			// update analytics config
			settingsAnalytics := &settingsTY.Settings{ID: settingsTY.KeyAnalytics}
			settingsAnalytics.Spec = utils.StructToMap(&settingsTY.AnalyticsConfig{AnonymousID: utils.RandUUID()})
			err = settingsAPI.UpdateSettings(settingsAnalytics)
			if err != nil {
				zap.L().Error("error on updating analytics config", zap.Error(err))
			}
		}

		// set everyday any time between 12:00 AM to 4:59 AM
		// get random hour and minute
		hour := rand.Intn(4)
		minute := rand.Intn(59)
		cronExpression := fmt.Sprintf("0 %d %d * * *", minute, hour)
		zap.L().Info("analytics data reporter job scheduled", zap.String("cron", cronExpression))

		helper.Schedule(analyticsJobName, cronExpression, analyticsAPI.ReportAnalyticsData)

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

	}
}
