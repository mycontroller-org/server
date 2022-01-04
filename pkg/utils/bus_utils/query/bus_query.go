package query

import (
	"context"
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

// QueryResource posts as request on response calls the callback
func QueryResource(resourceID, resourceType, command string, data interface{}, callBack func(item interface{}) bool, out interface{}, timeout time.Duration) error {
	return QueryService(mcbus.TopicServiceResourceServer, resourceID, resourceType, command, data, callBack, out, timeout)
}

// QueryService posts as request to a service, on response calls the callback
func QueryService(serviceTopic, resourceID, resourceType, command string, data interface{}, callBack func(item interface{}) bool, out interface{}, timeout time.Duration) error {
	// NOTE: when we have 'capacity' as '0', deadlocked somewhere, never returned from 'closeChan.SafeClose()'
	// for now changed the capacity to '1', and works as expected
	// TODO: find the blocker call and fix it
	closeChan := concurrency.NewChannel(1)
	defer closeChan.SafeClose()

	replyTopic := mcbus.FormatTopic(fmt.Sprintf("internal_query_response_%s", utils.RandIDWithLength(5)))
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

	busUtils.PostToService(serviceTopic, resourceID, data, resourceType, command, replyTopic)

	select {
	case <-closeChan.CH:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("reached timeout: %s", timeout.String())
	}
}

func responseFunc(closeChan *concurrency.Channel, callBack func(item interface{}) bool, out interface{}) func(data *busTY.BusData) {
	return func(busData *busTY.BusData) {
		event := &rsTY.ServiceEvent{}
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
