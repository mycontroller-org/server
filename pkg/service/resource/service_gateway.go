package resource

import (
	"errors"

	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	"github.com/mycontroller-org/server/v2/pkg/model"
	rsML "github.com/mycontroller-org/server/v2/pkg/model/resource_service"
	concurrencyUtils "github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

var gwReconnectStore = concurrencyUtils.NewStore()

func gatewayService(reqEvent *rsML.ServiceEvent) error {
	resEvent := &rsML.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsML.CommandGet:
		data, err := getGateway(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsML.CommandUpdateState:
		err := updateGatewayState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsML.CommandLoadAll:
		gatewayAPI.LoadAll()

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getGateway(request *rsML.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		gwConfig, err := gatewayAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return gwConfig, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := gatewayAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func updateGatewayState(reqEvent *rsML.ServiceEvent) error {
	if reqEvent.Data == nil {
		zap.L().Error("gateway state not supplied", zap.Any("event", reqEvent))
		return errors.New("gateway state not supplied")
	}

	state := &model.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	err = gatewayAPI.SetState(reqEvent.ID, state)
	if err != nil {
		return err
	}

	if state.Status == model.StatusUp {
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
	gw, err := gatewayAPI.GetByID(reqEvent.ID)
	if err != nil {
		return err
	}
	delay := gw.GetReconnectDelay()
	if gw.Enabled && delay != nil && !gwReconnectStore.IsAvailable(gw.ID) {
		job := concurrencyUtils.GetAsyncRunner(getTriggerGatewayStartFunc(reqEvent.ID), *delay, true)
		gwReconnectStore.Add(gw.ID, job)
		job.StartAsync()
	}

	return nil
}

func getTriggerGatewayStartFunc(gatewayID string) func() {
	return func() {
		gwReconnectStore.Remove(gatewayID)
		gw, err := gatewayAPI.GetByID(gatewayID)
		if err != nil {
			zap.L().Debug("error on getting gateway instance. may be deleted?", zap.String("gateway", gatewayID), zap.String("error", err.Error()))
			return
		}
		if !gw.Enabled || gw.State.Status == model.StatusUp {
			return
		}

		err = gatewayAPI.Reload([]string{gatewayID})
		if err != nil {
			zap.L().Error("error on reloading a gateway", zap.String("gateway", gatewayID), zap.Error(err))
		}
	}
}
