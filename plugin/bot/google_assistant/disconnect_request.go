package googleassistant

import (
	gaTY "github.com/mycontroller-org/server/v2/plugin/bot/google_assistant/types"
	"go.uber.org/zap"
)

func runDisconnectRequest(request gaTY.Request) {
	zap.L().Info("*** received disconnect request ***", zap.Any("request", request))

}
