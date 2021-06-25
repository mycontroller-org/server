package esphome

import (
	"fmt"
	"sync"

	gwML "github.com/mycontroller-org/server/v2/pkg/model/gateway"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

// default values
const (
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
	GatewayConfig *gwML.Config
	clientStore   *ClientStore
	entityStore   *EntityStore
	rxMessageFunc func(rawMsg *msgML.RawMessage) error
}

// Init provider
func Init(gatewayCfg *gwML.Config) (*Provider, error) {
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

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgML.RawMessage) error) error {
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
	p.unscheduleAll()
	p.clientStore.Close()
	p.entityStore.Close()
	return nil
}

// unscheduleAll removes all the schedules added by this gateway
func (p *Provider) unscheduleAll() {
	coreScheduler.SVC.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, p.GatewayConfig.ID))
}

// Unschedule removes a schedule
func Unschedule(scheduleID string) {
	coreScheduler.SVC.RemoveFunc(scheduleID)
}

// Schedule adds a schedule
func Schedule(schedulerID, interval string, triggerFunc func()) error {
	cronSpec := fmt.Sprintf("@every %s", interval)
	err := coreScheduler.SVC.AddFunc(schedulerID, cronSpec, triggerFunc)
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
		return err
	}
	zap.L().Debug("added a schedule", zap.String("schedulerID", schedulerID), zap.String("interval", interval))
	return nil
}
