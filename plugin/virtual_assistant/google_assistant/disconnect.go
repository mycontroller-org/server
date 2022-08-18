package google_assistant

import (
	gaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/google_assistant/types"
	"go.uber.org/zap"
)

func runDisconnectRequest(request gaTY.Request) {
	zap.L().Info("*** received disconnect request ***", zap.Any("request", request))

}
