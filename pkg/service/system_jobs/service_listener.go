package systemjobs

import (
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// Start service listener
func (svc *SystemJobsService) Start() error {
	// on message receive add it in to our local queue
	sID, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEvent)
	if err != nil {
		return err
	}
	svc.eventsQueue.SubscriptionId = sID

	// reload startup jobs
	svc.reloadSunriseJob()
	svc.reloadTelemetryJob()
	svc.reloadNodeStateVerifyJob()

	return nil
}

// Close the service listener
func (svc *SystemJobsService) Close() error {
	err := svc.bus.Unsubscribe(svc.eventsQueue.Topic, svc.eventsQueue.SubscriptionId)
	if err != nil {
		svc.logger.Error("error on unsubscribe", zap.String("topic", svc.eventsQueue.Topic), zap.Error(err))
	}
	svc.eventsQueue.Close()
	return nil
}

func (svc *SystemJobsService) onEvent(event *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := event.LoadData(reqEvent)
	if err != nil {
		svc.logger.Warn("failed to convert to target type", zap.Error(err))
		return
	}
	if reqEvent.Type == "" {
		svc.logger.Warn("received an empty event", zap.Any("event", event))
		return
	}
	svc.logger.Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := svc.eventsQueue.Produce(reqEvent)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func (svc *SystemJobsService) processEvent(event interface{}) error {
	reqEvent := event.(*rsTY.ServiceEvent)
	svc.logger.Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeSystemJobs {
		svc.logger.Warn("unsupported event type", zap.Any("event", reqEvent))
		return nil
	}

	if reqEvent.Command != rsTY.CommandReload {
		svc.logger.Warn("unsupported command", zap.Any("event", reqEvent))
		return nil
	}

	switch reqEvent.Data {

	case rsTY.SubCommandJobNodeStatusUpdater:
		svc.reloadNodeStateVerifyJob()

	case rsTY.SubCommandJobSunriseTimeUpdater:
		svc.reloadSunriseJob()

	default:
		// NOOP
	}
	return nil
}
