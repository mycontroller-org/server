package service

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	raml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gateway/provider"
	"go.uber.org/zap"
)

// Start gateway
func Start(gatewayCfg *gwml.Config) error {
	if gwService.Get(gatewayCfg.ID) != nil {
		return fmt.Errorf("A service is in running state. gateway:%s", gatewayCfg.ID)
	}
	if !gatewayCfg.Enabled { // this gateway is not enabled
		return nil
	}
	zap.L().Debug("Starting a gateway", zap.Any("id", gatewayCfg.ID))
	state := ml.State{Since: time.Now()}

	service, err := gwpd.GetService(gatewayCfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		zap.L().Error("Unable to start a gateway service", zap.Any("id", gatewayCfg.ID), zap.Error(err))
		state.Message = err.Error()
		state.Status = ml.StateDown
	} else {
		state.Message = "Started successfully"
		state.Status = ml.StateUp
		gwService.Add(service)
	}

	setState(gatewayCfg.ID, state)
	return nil
}

// Stop gateway
func Stop(id string) error {
	zap.L().Debug("Stopping a gateway", zap.Any("id", id))
	service := gwService.Get(id)
	if service != nil {
		err := service.Stop()
		state := ml.State{
			Status:  ml.StateDown,
			Since:   time.Now(),
			Message: "Stopped by request",
		}
		if err != nil {
			zap.L().Error("Failed to stop gateway service", zap.String("id", id), zap.Error(err))
			state.Message = err.Error()
		}
		setState(id, state)
		gwService.Remove(id)
	}
	return nil
}

// Reload gateway
func Reload(gwCfg *gwml.Config) error {
	err := Stop(gwCfg.ID)
	if err != nil {
		return err
	}
	return Start(gwCfg)
}

// UnloadAll stops all the gateways
func UnloadAll() {
	ids := gwService.ListIDs()
	for _, id := range ids {
		err := Stop(id)
		if err != nil {
			zap.L().Error("error on stopping a gateway", zap.String("id", id))
		}
	}
}

func setState(id string, state ml.State) {
	event := &raml.Event{
		Type:    raml.TypeGateway,
		Command: raml.CommandUpdateState,
		ID:      id,
	}
	event.SetData(state)
	topic := mcbus.FormatTopic(mcbus.TopicResourceServer)
	err := mcbus.Publish(topic, event)
	if err != nil {
		zap.L().Error("failed to post an event", zap.String("topic", topic), zap.Any("event", event))
	}
}
