package systemjobs

import (
	"fmt"
	"time"

	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	helper "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/helper_utils"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	idNodeStatusVerifier     = "node_status_verifier"
	DefaultExecutionInterval = "15m"
	DefaultInactiveDuration  = "15m"
)

func ReloadNodeStateVerifyJob() {
	// get interval and inactive delay
	settings, err := settingsAPI.GetSystemSettings()
	if err != nil {
		zap.L().Error("error on getting system settings", zap.Error(err))
		return
	}
	executionInterval := settings.NodeStateJob.ExecutionInterval
	if executionInterval == "" {
		executionInterval = DefaultExecutionInterval
	}

	inactiveDurationString := settings.NodeStateJob.InactiveDuration
	if inactiveDurationString == "" {
		inactiveDurationString = DefaultInactiveDuration
	}
	inactiveDuration := utils.ToDuration(inactiveDurationString, time.Minute*15)

	// func to call verify node status with inactive interval
	verifyNodeState := func() {
		nodeAPI.VerifyNodeUpStatus(inactiveDuration)
	}

	// schedule a job
	helper.Schedule(idNodeStatusVerifier, fmt.Sprintf("@every %s", executionInterval), verifyNodeState)
}
