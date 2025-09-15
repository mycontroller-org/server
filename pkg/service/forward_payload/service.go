package forwardpayload

import (
	"context"
	"fmt"

	actionAPI "github.com/mycontroller-org/server/v2/pkg/api/action"
	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	"github.com/mycontroller-org/server/v2/pkg/types/field"
	fedPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const (
	paginationLimit  = int64(50)
	defaultQueueSize = int(100)
	defaultWorkers   = int(1)
)

type ForwardPayloadService struct {
	logger      *zap.Logger
	api         *entityAPI.API
	actionApi   *actionAPI.ActionAPI
	bus         busTY.Plugin
	eventsQueue *queueUtils.QueueSpec
}

func New(ctx context.Context) (serviceTY.Service, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	actionAPI, err := actionAPI.New(ctx)
	if err != nil {
		return nil, err
	}

	svc := &ForwardPayloadService{
		logger:    logger.Named("forward_payload_service"),
		api:       api,
		bus:       bus,
		actionApi: actionAPI,
	}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Topic:          topic.TopicEventField,
		Queue:          queueUtils.New(svc.logger, "forward_payload", defaultQueueSize, svc.processEvent, defaultWorkers),
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *ForwardPayloadService) Name() string {
	return "forward_payload_service"
}

// Start message process engine
func (svc *ForwardPayloadService) Start() error {
	// on event receive add it in to local queue
	sID, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEventReceive)
	if err != nil {
		return err
	}

	svc.eventsQueue.SubscriptionId = sID
	return nil
}

func (svc *ForwardPayloadService) onEventReceive(busData *busTY.BusData) {
	event := &eventTY.Event{}
	err := busData.LoadData(event)
	if err != nil {
		svc.logger.Warn("Error on convert to target type", zap.Any("topic", busData.Topic), zap.Error(err))
		return
	}

	if event.EntityType != types.EntityField || event.Type != eventTY.TypeUpdated {
		// this data is not for us
		return
	}

	if event.Entity == nil {
		svc.logger.Warn("Received a nil data", zap.Any("busData", busData))
		return
	}

	field := field.Field{}
	err = event.LoadEntity(&field)
	if err != nil {
		svc.logger.Warn("error on conversion", zap.Any("entity", event), zap.Error(err))
		return
	}

	svc.logger.Debug("Field data added into processing queue", zap.Any("data", field))
	status := svc.eventsQueue.Produce(&field)
	if !status {
		svc.logger.Warn("error to store the data into queue", zap.Any("data", field))
	}
}

// Close message process engine
func (svc *ForwardPayloadService) Close() error {
	err := svc.bus.Unsubscribe(svc.eventsQueue.Topic, svc.eventsQueue.SubscriptionId)
	if err != nil {
		svc.logger.Error("error on unsubscription", zap.Error(err), zap.String("topic", svc.eventsQueue.Topic), zap.Int64("subscriptionId", svc.eventsQueue.SubscriptionId))
	}
	svc.eventsQueue.Close()
	return nil
}

// processEvent from the queue
func (svc *ForwardPayloadService) processEvent(item interface{}) error {
	field := item.(*field.Field)

	quickID, err := quickIdUtils.GetQuickID(*field)
	if err != nil {
		svc.logger.Error("unable to get quick id", zap.Error(err), zap.String("gateway", field.GatewayID), zap.String("node", field.NodeID), zap.String("source", field.SourceID), zap.String("field", field.FieldID))
		return nil
	}

	// fetch mapped filed for this event
	pagination := &storageTY.Pagination{Limit: 50}
	filters := []storageTY.Filter{
		{Key: types.KeySrcFieldID, Operator: storageTY.OperatorEqual, Value: quickID},
		{Key: types.KeyEnabled, Operator: storageTY.OperatorEqual, Value: true},
	}
	response, err := svc.api.ForwardPayload().List(filters, pagination)
	if err != nil {
		svc.logger.Error("error getting mapping data from database", zap.Error(err))
		return nil
	}

	if response.Count == 0 {
		return nil
	}

	svc.logger.Debug("Starting data forwarding", zap.Any("data", field))

	mappings := *response.Data.(*[]fedPayloadTY.Config)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		// send payload
		if mapping.SrcFieldID != mapping.DstFieldID {
			err = svc.actionApi.ToFieldByQuickID(mapping.DstFieldID, fmt.Sprintf("%v", field.Current.Value))
			if err != nil {
				svc.logger.Error("error on sending payload", zap.Any("mapping", mapping), zap.Error(err))
			} else {
				svc.logger.Debug("Data forwarded", zap.Any("mapping", mapping))
			}
		}
	}
	return nil
}
