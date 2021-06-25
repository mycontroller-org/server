package googleassistant

import (
	"github.com/mycontroller-org/server/v2/plugin/bot/google_assistant/model"
	"go.uber.org/zap"
)

func runDisconnectRequest(request model.Request) {
	zap.L().Info("*** received disconnect request ***", zap.Any("request", request))

}
