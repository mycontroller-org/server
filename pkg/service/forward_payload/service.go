package forwardpayload

import (
	"fmt"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	fpAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	"github.com/mycontroller-org/backend/v2/pkg/model/field"
	fpml "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

var (
	queue       *q.BoundedQueue
	queueSize   = int(1000)
	workerCount = int(1)
)

// Init message process engine
func Init() error {
	queue = utils.GetQueue("forward_payload", queueSize)

	// on event receive add it in to local queue
	_, err := mcbus.Subscribe(mcbus.TopicEventSensorFieldSet, onEventReceive)
	if err != nil {
		return err
	}

	queue.StartConsumers(workerCount, processEvent)
	return nil
}

func onEventReceive(event *event.Event) {
	field := &field.Field{}
	err := event.ToStruct(field)
	if err != nil {
		zap.L().Warn("Error on convet to target type", zap.Error(err))
		return
	}

	if field == nil {
		zap.L().Warn("Received a nil data", zap.Any("event", event))
		return
	}
	zap.L().Debug("Sensor Field data added into processing queue", zap.Any("data", field))
	status := queue.Produce(field)
	if !status {
		zap.L().Warn("error to store the data into queue", zap.Any("data", field))
	}
}

// Close message process engine
func Close() error {
	queue.Stop()
	return nil
}

// processEvent from the queue
func processEvent(item interface{}) {
	field := item.(*field.Field)

	// fetch mapped filed for this event
	pagination := &stgml.Pagination{Limit: 50}
	filters := []stgml.Filter{
		{Key: ml.KeySourceID, Operator: stgml.OperatorEqual, Value: field.ID},
		{Key: ml.KeyEnabled, Operator: stgml.OperatorEqual, Value: true},
	}
	response, err := fpAPI.List(filters, pagination)
	if err != nil {
		zap.L().Error("error getting mapping data from database", zap.Error(err))
		return
	}

	if response.Count == 0 {
		return
	}

	zap.L().Debug("Starting data forwarding", zap.Any("data", field))

	mappings := *response.Data.(*[]fpml.Mapping)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		// send payload
		if mapping.SourceID != mapping.TargetID {
			err = action.ToSensorField(mapping.TargetID, fmt.Sprintf("%v", field.Payload.Value))
			if err != nil {
				zap.L().Error("error on sending payload", zap.Any("mapping", mapping), zap.Error(err))
			} else {
				zap.L().Debug("Data forwarded", zap.Any("mapping", mapping))
			}
		}
	}
}
