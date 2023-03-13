package systemjobs

import (
	"fmt"
	"math/rand"

	telemetryAPI "github.com/mycontroller-org/server/v2/pkg/telemetry"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	utils "github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	idTelemetryJob = "telemetry"
)

// ReloadJob func
func (svc *SystemJobsService) reloadTelemetryJob() {
	// update telemetry job
	jobName := svc.getScheduleID(idTelemetryJob)
	svc.unschedule(jobName)
	if types.GetEnvBool(types.ENV_TELEMETRY_ENABLED) {
		// generate telemetryId if not available
		_, err := svc.api.Settings().GetTelemetry()
		if err != nil {
			// update telemetry config
			settingsTelemetry := &settingsTY.Settings{ID: settingsTY.KeyTelemetry}
			settingsTelemetry.Spec = utils.StructToMap(&settingsTY.TelemetryConfig{AnonymousID: utils.RandUUID()})
			err = svc.api.Settings().UpdateSettings(settingsTelemetry)
			if err != nil {
				svc.logger.Error("error on updating telemetry config", zap.Error(err))
			}
		}

		// set everyday any time between 12:00 AM to 4:59 AM
		// get random hour and minute
		hour := rand.Intn(4)
		minute := rand.Intn(59)
		cronExpression := fmt.Sprintf("0 %d %d * * *", minute, hour)
		svc.logger.Info("telemetry data reporter job scheduled", zap.String("cron", cronExpression))

		svc.schedule(jobName, cronExpression, svc.telemetryCallBack())

		svc.logger.Info(`
#=========================================================================================================#
#                                   Non-PII data Processing Agreement                                     #
#                                   ---------------------------------                                     #
# (non-PII - non Personally Identifiable Information)                                                     #
# Telemetry service is enabled in your server, says you agreed that the non-PII data will be collected,   #
# processed, and used by MyController.org to improve the quality of the software.                         #
#                                                                                                         #
# If you do not like to share the non-PII data, you can disable the telemetry service via configuration   #
# file and restart the server.                                                                            #
#                                                                                                         #
# file: mycontroller.yaml                                                                                 #
# telemetry:                                                                                              #
#   enabled: false                                                                                        #
#=========================================================================================================#
`)

	}
}

func (svc *SystemJobsService) telemetryCallBack() func() {
	return func() {
		telemetryAPI.ReportTelemetryData(svc.ctx)
	}
}
