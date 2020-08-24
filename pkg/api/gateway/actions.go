package gateway

import (
	"time"

	gm "github.com/mycontroller-org/backend/pkg/gateway"
	gms "github.com/mycontroller-org/backend/pkg/gateway/serial"
	ml "github.com/mycontroller-org/backend/pkg/model"
	gwml "github.com/mycontroller-org/backend/pkg/model/gateway"
	svc "github.com/mycontroller-org/backend/pkg/service"
	"github.com/mycontroller-org/backend/plugin/gateway_provider/mysensors"
	"go.uber.org/zap"
)

// Start gateway
func Start(g *gwml.Config) error {

	var parser gwml.MessageParser

	switch g.Provider.Type {
	case gwml.ProviderMySensors:
		parser = &mysensors.Parser{Gateway: g}
		// update serial message splitter
		g.Provider.Config[gms.KeyMessageSplitter] = mysensors.SerialMessageSplitter
	default:
		zap.L().Info("Unknown provider", zap.Any("gateway", g))
		return nil
	}
	s := &gwml.Service{
		Config: g,
		Parser: parser,
	}

	err := gm.Start(s)
	if err != nil {
		zap.L().Info("Unable to start the gateway", zap.Any("gateway", g))
	} else {
		svc.AddGatewayService(s)
	}

	return nil
}

// Stop gateway
func Stop(g *gwml.Config) error {
	gs := svc.GetGatewayService(g)
	if gs != nil {
		err := gm.Stop(gs)
		if err != nil {
			zap.L().Error("Failed to stop media service", zap.Error(err))
		}
	}
	return nil
}

// Reload gateway
func Reload(g *gwml.Config) error {
	err := Stop(g)
	if err != nil {
		return err
	}
	if g.Enabled {
		err = Start(g)
	}
	return err
}

// LoadGateways makes gateways alive
func LoadGateways() {
	gws, err := ListGateways(nil, ml.Pagination{})
	if err != nil {
		zap.L().Error("Error getting list of gateways", zap.Error(err))
	}
	for _, gw := range gws {
		state := ml.State{
			Since:   time.Now(),
			Status:  ml.StateUp,
			Message: "Started successfully",
		}
		err = Start(&gw)
		if err != nil {
			state.Message = err.Error()
			state.Status = ml.StateDown
			zap.L().Error("Error loading a gateway", zap.Error(err), zap.Any("gateway", gw))
		}
		err = SetState(&gw, state)
		if err != nil {
			zap.L().Error("Failed to update gateway status")
		}
	}
}

// UnloadGateways makes stop gateways
func UnloadGateways() {
	gws, err := ListGateways(nil, ml.Pagination{})
	if err != nil {
		zap.L().Error("Error getting list of gateways", zap.Error(err))
	}
	for _, gw := range gws {
		err = Stop(&gw)
		if err != nil {
			zap.L().Error("Error unloading a gateway", zap.Error(err), zap.Any("gateway", gw))
		}
	}
}
