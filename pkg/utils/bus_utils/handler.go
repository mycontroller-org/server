package busutils

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// PostToHandler send data to handlers
func PostToHandler(logger *zap.Logger, bus busTY.Plugin, handlers []string, parameters map[string]interface{}) {
	logger.Debug("posting data to handlers", zap.Any("handlers", handlers))

	// remove disabled parameters
	updateData := make(map[string]interface{})
	for name, rawParameter := range parameters {
		parameterMap, ok := rawParameter.(map[string]interface{})
		if !ok {
			logger.Warn("received parameter is not a map[string]interface{}", zap.String("parameterName", name), zap.String("type", fmt.Sprintf("%T", rawParameter)))
			continue
		}
		parameter := cmap.CustomMap(parameterMap)
		if parameter.GetBool(types.KeyDisabled) {
			continue
		}

		// update to our new list if the item is not disabled
		updateData[name] = parameter
	}

	if len(updateData) == 0 {
		return
	}

	for _, handlerID := range handlers {
		if handlerID == "" {
			continue
		}
		msg := &handlerType.MessageWrapper{
			ID:   handlerID,
			Data: updateData,
		}
		err := bus.Publish(topic.TopicPostMessageNotifyHandler, msg)
		if err != nil {
			logger.Error("error on posting data to handler", zap.Error(err), zap.String("handlerID", handlerID))
		}
	}
}
