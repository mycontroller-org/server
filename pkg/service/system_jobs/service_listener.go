package systemjobs

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	helper "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/helper_utils"
	nodeJob "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/node_job"
	sunriseJob "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/sunrise_job"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	serviceMessageQueueLimit = 50
	serviceMessageQueueName  = "service_listener_system_jobs"
)

var (
	serviceQueue   *queueUtils.Queue
	topic          = mcbus.TopicInternalSystemJobs
	subscriptionID = int64(-1)
)

// Start service listener
func Start() error {
	serviceQueue = queueUtils.New(serviceMessageQueueName, serviceMessageQueueLimit, processEvent, 1)

	// on message receive add it in to our local queue
	sID, err := mcbus.Subscribe(mcbus.FormatTopic(topic), onEvent)
	subscriptionID = sID
	return err
}

// Close the service listener
func Close() {
	err := mcbus.Unsubscribe(mcbus.FormatTopic(topic), subscriptionID)
	if err != nil {
		zap.L().Error("error on unsubscribe", zap.String("topic", mcbus.FormatTopic(topic)), zap.Error(err))
	}
	serviceQueue.Close()
}

func onEvent(event *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := event.LoadData(reqEvent)
	if err != nil {
		zap.L().Warn("failed to convet to target type", zap.Error(err))
		return
	}
	if reqEvent == nil {
		zap.L().Warn("received a nil message", zap.Any("event", event))
		return
	}
	zap.L().Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := serviceQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func processEvent(event interface{}) {
	reqEvent := event.(*rsTY.ServiceEvent)
	zap.L().Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeSystemJobs {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
		return
	}

	if reqEvent.Command != rsTY.CommandReload {
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
		return
	}

	switch reqEvent.Data {

	case helper.JobTypeNodeStateUpdater:
		nodeJob.ReloadNodeStateVerifyJob()

	case helper.JobTypeSunriseUpdater:
		sunriseJob.ReloadJob()

	default:
		// NOOP
	}
}
