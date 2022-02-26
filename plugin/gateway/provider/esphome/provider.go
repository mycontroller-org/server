package esphome

import (
	"fmt"
	"sync"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	scheduleUtils "github.com/mycontroller-org/server/v2/pkg/utils/schedule"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

// default values
const (
	PluginEspHome = "esphome"

	defaultTimeout            = "5s"
	defaultAliveCheckInterval = "15s"

	schedulePrefix = "esphome_gw"
)

// Config data for this gateway
type Config struct {
	Password           string
	Timeout            string
	AliveCheckInterval string
	Nodes              map[string]ESPHomeNodeConfig
}

// Provider data
type Provider struct {
	Config        Config
	GatewayConfig *gwTY.Config
	clientStore   *ClientStore
	entityStore   *EntityStore
	rxMessageFunc func(rawMsg *msgTY.RawMessage) error
}

// NewPluginEspHome provider
func NewPluginEspHome(gatewayCfg *gwTY.Config) (providerTY.Plugin, error) {
	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameNone, gatewayCfg.Provider, &cfg)
	if err != nil {
		return nil, err
	}

	// verify and update defaults
	cfg.Timeout = utils.ValidDuration(cfg.Timeout, defaultTimeout)
	cfg.AliveCheckInterval = utils.ValidDuration(cfg.AliveCheckInterval, defaultAliveCheckInterval)

	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayCfg,
		clientStore:   &ClientStore{nodes: make(map[string]*ESPHomeNode), mutex: &sync.RWMutex{}},
		entityStore:   &EntityStore{nodes: make(map[string]map[uint32]Entity), mutex: &sync.RWMutex{}},
		rxMessageFunc: nil,
	}
	zap.L().Debug("Config details", zap.Any("received", gatewayCfg.Provider), zap.Any("converted", cfg))
	return provider, nil
}

func (p *Provider) Name() string {
	return PluginEspHome
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgTY.RawMessage) error) error {
	// update receive message function
	p.rxMessageFunc = rxMessageFunc

	// create espnode clients
	for nodeID, nodeCfg := range p.Config.Nodes {
		if nodeCfg.Disabled {
			continue
		}
		zap.L().Debug("connecting to node", zap.Any("gatewayId", p.GatewayConfig.ID), zap.Any("nodeID", nodeID))

		if nodeCfg.Password == "" {
			nodeCfg.Password = p.Config.Password
		}

		nodeCfg.Timeout = utils.ValidDuration(nodeCfg.Timeout, p.Config.Timeout)
		nodeCfg.AliveCheckInterval = utils.ValidDuration(nodeCfg.AliveCheckInterval, p.Config.AliveCheckInterval)
		nodeCfg.ReconnectDelay = utils.ValidDuration(nodeCfg.ReconnectDelay, p.GatewayConfig.ReconnectDelay)

		espNode := NewESPHomeNode(p.GatewayConfig.ID, nodeID, nodeCfg, p.entityStore, p.rxMessageFunc)

		err := espNode.Connect()
		if err != nil {
			zap.L().Info("error on connecting a node", zap.String("gatewayId", espNode.GatewayID), zap.String("nodeId", nodeID), zap.String("error", err.Error()))
			espNode.ScheduleReconnect()
		}

		p.clientStore.AddNode(nodeID, espNode)
	}

	return nil
}

// Close func
func (p *Provider) Close() error {
	scheduleUtils.UnscheduleAll(schedulePrefix, p.GatewayConfig.ID)
	p.clientStore.Close()
	p.entityStore.Close()
	return nil
}

// Unschedule removes a schedule
func Unschedule(scheduleID string) {
	scheduleUtils.Unschedule(scheduleID)
}

// Schedule adds a schedule
func Schedule(scheduleID, interval string, triggerFunc func()) error {
	jobSpec := fmt.Sprintf("@every %s", interval)
	return scheduleUtils.Schedule(scheduleID, jobSpec, triggerFunc)
}
