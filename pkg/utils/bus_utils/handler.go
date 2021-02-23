package busutils

import (
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"go.uber.org/zap"
)

// PostToHandler send data to handlers
func PostToHandler(handlers []string, data map[string]interface{}) {
	zap.L().Debug("Posting data to handlers", zap.Any("handlers", handlers))
	for _, handlerID := range handlers {
		msg := &handlerML.MessageWrapper{
			ID:   handlerID,
			Data: data,
		}
		err := mcbus.Publish(mcbus.FormatTopic(mcbus.TopicPostMessageNotifyHandler), msg)
		if err != nil {
			zap.L().Error("error on posting data to handler", zap.Error(err), zap.String("handlerID", handlerID))
		}
	}
}
