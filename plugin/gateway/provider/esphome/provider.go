package esphome

import (
	"context"
	"fmt"
	"sync"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
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

	loggerName = "gateway_esphome"
)

// Config data for this gateway
type Config struct {
	Password           string
	EncryptionKey      string
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
	logger        *zap.Logger
	scheduler     schedulerTY.CoreScheduler
	bus           busTY.Plugin
}

// esphome provider
func New(ctx context.Context, config *gwTY.Config) (providerTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	scheduler, err := schedulerTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	err = utils.MapToStruct(utils.TagNameNone, config.Provider, &cfg)
	if err != nil {
		return nil, err
	}

	// verify and update defaults
	cfg.Timeout = utils.ValidDuration(cfg.Timeout, defaultTimeout)
	cfg.AliveCheckInterval = utils.ValidDuration(cfg.AliveCheckInterval, defaultAliveCheckInterval)

	namedLogger := logger.Named(loggerName)

	provider := &Provider{
		Config:        cfg,
		GatewayConfig: config,
		clientStore:   &ClientStore{nodes: make(map[string]*ESPHomeNode), mutex: &sync.RWMutex{}, logger: namedLogger},
		entityStore:   &EntityStore{nodes: make(map[string]map[uint32]Entity), mutex: &sync.RWMutex{}},
		rxMessageFunc: nil,
		logger:        namedLogger,
		scheduler:     scheduler,
		bus:           bus,
	}
	provider.logger.Debug("Config details", zap.Any("received", config.Provider), zap.Any("converted", cfg))
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
		p.logger.Debug("connecting to node", zap.Any("gatewayId", p.GatewayConfig.ID), zap.Any("nodeID", nodeID))

		if nodeCfg.UseGlobalPassword {
			nodeCfg.Password = p.Config.Password
		}

		if nodeCfg.UseGlobalEncryptionKey {
			nodeCfg.EncryptionKey = p.Config.EncryptionKey
		}

		nodeCfg.Timeout = utils.ValidDuration(nodeCfg.Timeout, p.Config.Timeout)
		nodeCfg.AliveCheckInterval = utils.ValidDuration(nodeCfg.AliveCheckInterval, p.Config.AliveCheckInterval)
		nodeCfg.ReconnectDelay = utils.ValidDuration(nodeCfg.ReconnectDelay, p.GatewayConfig.ReconnectDelay)

		espNode := NewESPHomeNode(p.logger, p.GatewayConfig.ID, nodeID, nodeCfg, p.entityStore, p.rxMessageFunc, p.scheduler, p.bus)

		err := espNode.Connect()
		if err != nil {
			p.logger.Info("error on connecting a node", zap.String("gatewayId", espNode.GatewayID), zap.String("nodeId", nodeID), zap.String("error", err.Error()))
			espNode.ScheduleReconnect()
		}

		p.clientStore.AddNode(nodeID, espNode)
	}

	return nil
}

// Close func
func (p *Provider) Close() error {
	p.scheduler.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, p.GatewayConfig.ID))
	p.clientStore.Close()
	p.entityStore.Close()
	return nil
}

// Unschedule removes a schedule
func (p *Provider) Unschedule(scheduleID string) {
	p.scheduler.RemoveFunc(scheduleID)
}

// Schedule adds a schedule
func (p *Provider) Schedule(scheduleID, interval string, triggerFunc func()) error {
	jobSpec := fmt.Sprintf("@every %s", interval)
	return p.scheduler.AddFunc(scheduleID, jobSpec, triggerFunc)
}
