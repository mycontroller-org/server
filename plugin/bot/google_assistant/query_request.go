package googleassistant

import (
	"github.com/mycontroller-org/backend/v2/plugin/bot/google_assistant/model"
	"go.uber.org/zap"
)

func runQueryRequest(request model.QueryRequest) *model.QueryResponse {
	zap.L().Info("received a query request", zap.Any("request", request))
	return nil
}
