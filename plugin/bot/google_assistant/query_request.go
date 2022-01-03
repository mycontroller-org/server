package googleassistant

import (
	gaTY "github.com/mycontroller-org/server/v2/plugin/bot/google_assistant/types"
	"go.uber.org/zap"
)

func runQueryRequest(request gaTY.QueryRequest) *gaTY.QueryResponse {
	zap.L().Info("received a query request", zap.Any("request", request))
	return nil
}
