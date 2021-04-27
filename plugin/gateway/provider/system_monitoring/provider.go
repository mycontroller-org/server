package systemmonitoring

import (
	gwML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/plugin/gateway/provider/system_monitoring/config"
	"go.uber.org/zap"
)

const (
	schedulePrefix = "schedule_system_monitoring_gw_"
)

// Provider data
type Provider struct {
	Config        config.Config
	HostConfig    *config.HostConfig
	GatewayConfig *gwML.Config
	NodeID        string
}

// Init MySensors provider
func Init(gatewayCfg *gwML.Config) (*Provider, error) {
	cfg := config.Config{}
	err := utils.MapToStruct(utils.TagNameNone, gatewayCfg.Provider, &cfg)
	if err != nil {
		return nil, err
	}

	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayCfg,
	}
	zap.L().Debug("Config details", zap.Any("received", gatewayCfg.Provider), zap.Any("converted", cfg))
	return provider, nil
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgml.RawMessage) error) error {
	// update node id
	nodeID, err := p.HostID()
	if err != nil {
		return err
	}
	p.NodeID = nodeID

	// update nodeID from config data
	if definedID, ok := p.Config.HostIDMap[nodeID]; ok {
		p.NodeID = definedID
	}

	// get this host config
	if hCfg, ok := p.Config.HostConfigMap[nodeID]; ok {
		p.HostConfig = &hCfg
	} else if hCfg, ok := p.Config.HostConfigMap[p.NodeID]; ok {
		p.HostConfig = &hCfg
	} else {
		p.HostConfig = &config.HostConfig{}
	}

	// update schedules
	if !p.HostConfig.Disabled {
		err = p.scheduleMonitoring()
		if err != nil {
			p.unloadAll()
			return err
		}
	}

	// post node details immediately
	p.updateNodeDetails()

	return nil
}

// Close func
func (p *Provider) Close() error {
	// do internal works
	p.unloadAll()
	return nil
}

// Post func
func (p *Provider) Post(rawMsg *msgml.RawMessage) error {
	// this gateway do not support post messages
	return nil
}
