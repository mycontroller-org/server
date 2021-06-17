package init

import (
	"net/http"

	"github.com/mycontroller-org/backend/v2/cmd/server/app/handler"
	"github.com/mycontroller-org/backend/v2/cmd/server/listener"
	"go.uber.org/zap"
)

var HANDLER http.Handler

func StartWebHandler() {
	if HANDLER == nil {
		httpHandler, err := handler.GetHandler()
		if err != nil {
			zap.L().Fatal("Error on getting handler", zap.Error(err))
		}
		HANDLER = httpHandler
		listener.StartListener(httpHandler)
		return
	}
	zap.L().Info("handler init service is called multiple times")
}
