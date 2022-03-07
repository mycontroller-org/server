package systemmonitoring

import (
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/plugin/gateway/provider/system_monitoring/config"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
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
	GatewayConfig *gwTY.Config
	NodeID        string
}

// NewPluginSystemMonitoring provider
func NewPluginSystemMonitoring(gatewayCfg *gwTY.Config) (providerTY.Plugin, error) {
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
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgTY.RawMessage) error) error {
	// update node id
	hostID, err := p.HostID()
	if err != nil {
		return err
	}

	nodeID := hostID

	// get node configuration based on host id
	for nodeIDkey, nodeCFG := range p.Config.HostConfigMap {
		if utils.ContainsString(nodeCFG.HostIDs, hostID) {
			nodeID = nodeIDkey
		}
	}

	p.NodeID = nodeID

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
