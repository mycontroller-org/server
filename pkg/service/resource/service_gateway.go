package resource

import (
	"errors"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	concurrencyUtils "github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

var gwReconnectStore = concurrencyUtils.NewStore()

func (svc *ResourceService) gatewayService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getGateway(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := svc.updateGatewayState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsTY.CommandLoadAll:
		svc.api.Gateway().LoadAll()

	default:
		return errors.New("unknown command")
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getGateway(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		gwConfig, err := svc.api.Gateway().GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return gwConfig, nil
	} else if len(request.Labels) > 0 {
		filters := svc.getLabelsFilter(request.Labels)
		result, err := svc.api.Gateway().List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func (svc *ResourceService) updateGatewayState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		svc.logger.Error("gateway state not supplied", zap.Any("event", reqEvent))
		return errors.New("gateway state not supplied")
	}

	state := &types.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	err = svc.api.Gateway().SetState(reqEvent.ID, state)
	if err != nil {
		return err
	}

	if state.Status == types.StatusUp {
		if gwReconnectStore.IsAvailable(reqEvent.ID) {
			jobInterface := gwReconnectStore.Get(reqEvent.ID)
			if jobInterface != nil {
				job, ok := jobInterface.(*concurrencyUtils.Runner)
				if ok {
					job.Close()
				}
			}
			gwReconnectStore.Remove(reqEvent.ID)
		}
		return nil
	}

	// if the gateway reports status not as UP and is in enabled state
	// should restart the gateway after the defined reconnect delay
	gw, err := svc.api.Gateway().GetByID(reqEvent.ID)
	if err != nil {
		return err
	}
	delay := gw.GetReconnectDelay()
	if gw.Enabled && delay != nil && !gwReconnectStore.IsAvailable(gw.ID) {
		job := concurrencyUtils.GetAsyncRunner(svc.getTriggerGatewayStartFunc(reqEvent.ID), *delay, true)
		gwReconnectStore.Add(gw.ID, job)
		job.StartAsync()
	}

	return nil
}

func (svc *ResourceService) getTriggerGatewayStartFunc(gatewayID string) func() {
	return func() {
		gwReconnectStore.Remove(gatewayID)
		gw, err := svc.api.Gateway().GetByID(gatewayID)
		if err != nil {
			svc.logger.Debug("error on getting gateway instance. may be deleted?", zap.String("gateway", gatewayID), zap.String("error", err.Error()))
			return
		}
		if !gw.Enabled || gw.State.Status == types.StatusUp {
			return
		}

		err = svc.api.Gateway().Reload([]string{gatewayID})
		if err != nil {
			svc.logger.Error("error on reloading a gateway", zap.String("gateway", gatewayID), zap.Error(err))
		}
	}
}
