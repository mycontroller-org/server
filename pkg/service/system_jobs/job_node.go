package systemjobs

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	idNodeStatusVerifier     = "node_status_verifier"
	DefaultExecutionInterval = "15m"
	DefaultInactiveDuration  = "15m"
)

func (svc *SystemJobsService) reloadNodeStateVerifyJob() {
	// get interval and inactive delay
	settings, err := svc.api.Settings().GetSystemSettings()
	if err != nil {
		svc.logger.Error("error on getting system settings", zap.Error(err))
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
		svc.api.Node().VerifyNodeUpStatus(inactiveDuration)
	}

	// schedule a job
	svc.schedule(idNodeStatusVerifier, fmt.Sprintf("@every %s", executionInterval), verifyNodeState)
}
