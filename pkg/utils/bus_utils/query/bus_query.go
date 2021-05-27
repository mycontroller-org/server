package query

import (
	"context"
	"fmt"
	"time"

	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// QueryResource posts as request on response calls the callback
func QueryResource(resourceID, resourceType, command string, data interface{}, callBack func(item interface{}) bool, out interface{}, timeout time.Duration) error {
	closeChan := concurrency.NewChannel(0)
	defer closeChan.SafeClose()

	replyTopic := mcbus.FormatTopic(fmt.Sprintf("query_response_%s", utils.RandIDWithLength(5)))
	sID, err := mcbus.Subscribe(replyTopic, responseFunc(closeChan, callBack, out))
	if err != nil {
		return err
	}

	defer func() {
		err := mcbus.Unsubscribe(replyTopic, sID)
		if err != nil {
			zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", replyTopic))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	busUtils.PostToResourceService(resourceID, data, resourceType, command, replyTopic)

	select {
	case <-closeChan.CH:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("reached timeout: %s", timeout.String())
	}
}

func responseFunc(closeChan *concurrency.Channel, callBack func(item interface{}) bool, out interface{}) func(data *busML.BusData) {
	return func(busData *busML.BusData) {
		event := &rsML.ServiceEvent{}
		err := busData.LoadData(event)
		if err != nil {
			zap.L().Error("error on converting to event type", zap.Error(err))
			closeChan.SafeSend(true)
			return
		}

		if event.Error != "" {
			zap.L().Error("error on response", zap.Any("response", event))
			closeChan.SafeSend(true)
			return
		}

		if callBack != nil {
			// convert data to callback type
			err := event.LoadData(out)
			if err != nil {
				zap.L().Error("error on converting to target type", zap.Error(err), zap.Any("event", event))
				closeChan.SafeSend(true)
				return
			}

			if !callBack(out) { // continue?
				closeChan.SafeSend(true)
				return
			}
			return
		}

		closeChan.SafeSend(true)
	}
}
