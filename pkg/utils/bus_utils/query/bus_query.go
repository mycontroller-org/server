package query

import (
	"context"
	"fmt"
	"time"

	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// QueryResource posts as request on response calls the callback
func QueryResource(logger *zap.Logger, bus busTY.Plugin, resourceID, resourceType, command string, data interface{}, callBack func(item interface{}) bool, out interface{}, timeout time.Duration) error {
	return QueryService(logger, bus, topic.TopicServiceResourceServer, resourceID, resourceType, command, data, callBack, out, timeout)
}

// QueryService posts as request to a service, on response calls the callback
func QueryService(logger *zap.Logger, bus busTY.Plugin, serviceTopic, resourceID, resourceType, command string, data interface{}, callBack func(item interface{}) bool, out interface{}, timeout time.Duration) error {
	// NOTE: when we receive response from multiple gateways, deadlock happens when the capacity is '0'
	// for now changed the capacity to '20', it means supports up to 20 listening gateways
	// TODO: This is not a right fix. introduce a permanent fix
	closeChan := concurrency.NewChannel(1)
	defer closeChan.SafeClose()

	replyTopic := fmt.Sprintf("internal_query_response_%s", utils.RandIDWithLength(5))
	sID, err := bus.Subscribe(replyTopic, responseFunc(logger, closeChan, callBack, out))
	if err != nil {
		return err
	}

	defer func() {
		err := bus.Unsubscribe(replyTopic, sID)
		if err != nil {
			logger.Error("error on unsubscribe", zap.Error(err), zap.String("topic", replyTopic))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	busUtils.PostToService(logger, bus, serviceTopic, resourceID, data, resourceType, command, replyTopic)

	select {
	case <-closeChan.CH:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("reached timeout: %s", timeout.String())
	}
}

func responseFunc(logger *zap.Logger, closeChan *concurrency.Channel, callBack func(item interface{}) bool, out interface{}) func(data *busTY.BusData) {
	return func(busData *busTY.BusData) {
		event := &rsTY.ServiceEvent{}
		err := busData.LoadData(event)
		if err != nil {
			logger.Error("error on converting to event type", zap.Error(err))
			closeChan.SafeSend(true)
			return
		}

		if event.Error != "" {
			logger.Error("error on response", zap.Any("response", event))
			closeChan.SafeSend(true)
			return
		}

		if callBack != nil {
			// convert data to callback type
			err := event.LoadData(out)
			if err != nil {
				logger.Error("error on converting to target type", zap.Error(err), zap.Any("event", event))
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
