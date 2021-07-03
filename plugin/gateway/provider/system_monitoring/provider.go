package systemmonitoring

import (
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/plugin/gateway/provider/system_monitoring/config"
	providerType "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwType "github.com/mycontroller-org/server/v2/plugin/gateway/type"
	"go.uber.org/zap"
)

const (
	PluginSystemMonitoring = "system_monitoring"

	schedulePrefix = "schedule_system_monitoring_gw_"
)

// Provider data
type Provider struct {
	Config        config.Config
	HostConfig    *config.HostConfig
	GatewayConfig *gwType.Config
	NodeID        string
}

// NewPluginSystemMonitoring provider
func NewPluginSystemMonitoring(gatewayCfg *gwType.Config) (providerType.Plugin, error) {
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

func (p *Provider) Name() string {
	return PluginSystemMonitoring
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgML.RawMessage) error) error {
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
