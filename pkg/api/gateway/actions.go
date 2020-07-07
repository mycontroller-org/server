package gateway

import (
	gm "github.com/mycontroller-org/mycontroller-v2/pkg/gateway"
	gms "github.com/mycontroller-org/mycontroller-v2/pkg/gateway/serial"
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	srv "github.com/mycontroller-org/mycontroller-v2/pkg/service"
	"github.com/mycontroller-org/mycontroller-v2/plugin/gateway_provider/mysensors"
	"go.uber.org/zap"
)

// Start gateway
func Start(g *ml.GatewayConfig) error {

	var parser ml.GatewayMessageParser

	switch g.Provider.Type {
	case ml.GatewayProviderMySensors:
		parser = &mysensors.Parser{Gateway: g}
		// update serial message splitter
		g.Provider.Config[gms.KeyMessageSplitter] = mysensors.SerialMessageSplitter
	default:
		zap.L().Info("Unknown provider", zap.Any("gateway", g))
		return nil
	}
	s := &ml.GatewayService{
		Config: g,
		Parser: parser,
	}

	err := gm.Start(s)
	if err != nil {
		zap.L().Info("Unable to start the gateway", zap.Any("gateway", g))
	} else {
		srv.AddGatewayService(s)
	}

	return nil
}

// Stop gateway
func Stop(g *ml.GatewayConfig) error {
	gs := srv.GetGatewayService(g)
	if gs != nil {
		err := gm.Stop(gs)
		if err != nil {
			zap.L().Error("Failed to stop media service", zap.Error(err))
		}
	}
	return nil
}

// Reload gateway
func Reload(g *ml.GatewayConfig) error {
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
		err = Start(&gw)
		if err != nil {
			zap.L().Error("Error loading a gateway", zap.Error(err), zap.Any("gateway", gw))
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
