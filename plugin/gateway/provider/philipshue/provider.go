package philipshue

import (
	"context"
	"fmt"
	"time"

	"github.com/amimof/huego"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const PluginPhilipsHue = "philips_hue"

const (
	loggerName                = "gateway_philipshue"
	schedulePrefix            = "philipshue_status"
	scheduleFormatSync        = "%s_sync"
	scheduleFormatBridge      = "%s_bridge"
	defaultSyncInterval       = "15m"
	defaultBridgeSyncInterval = "10m"
)

// Config data
type Config struct {
	Host               string
	Username           string
	SyncInterval       string
	BridgeSyncInterval string
}

// Provider data
type Provider struct {
	ctx           context.Context
	Config        Config
	GatewayConfig *gwTY.Config
	bridge        *huego.Bridge
	logger        *zap.Logger
	scheduler     schedulerTY.CoreScheduler
	bus           busTY.Plugin
}

// philipshue provider
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

	namedLogger := logger.Named(loggerName)

	_, err = time.ParseDuration(cfg.SyncInterval)
	if err != nil {
		namedLogger.Debug("invalid sync interval supplied. set to default", zap.String("input", cfg.SyncInterval), zap.String("default", defaultSyncInterval))
		cfg.SyncInterval = defaultSyncInterval
	}

	// using sync insterval for bridge sync too
	cfg.BridgeSyncInterval = cfg.SyncInterval

	_, err = time.ParseDuration(cfg.BridgeSyncInterval)
	if err != nil {
		namedLogger.Debug("invalid bridge sync interval supplied. set to default", zap.String("input", cfg.BridgeSyncInterval), zap.String("default", defaultBridgeSyncInterval))
		cfg.BridgeSyncInterval = defaultBridgeSyncInterval
	}

	provider := &Provider{
		ctx:           ctx,
		Config:        cfg,
		GatewayConfig: config,
		logger:        namedLogger,
		scheduler:     scheduler,
		bus:           bus,
	}
	provider.logger.Debug("Config details", zap.Any("received", config.Provider), zap.Any("converted", cfg))
	return provider, nil
}

func (p *Provider) Name() string {
	return PluginPhilipsHue
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgTY.RawMessage) error) error {
	bridge := huego.New(p.Config.Host, p.Config.Username)
	_, err := bridge.GetLights()
	if err != nil {
		return err
	}
	p.bridge = bridge

	// schedules
	p.unscheduleAll() // removes the existing schedule, if any

	err = p.scheduleSync()
	if err != nil {
		return err
	}
	err = p.scheduleBridgeSync()
	if err != nil {
		return err
	}

	// update bridge details
	p.updateBridgeDetails()

	// on startup sync the status
	p.updateLights()
	p.updateSensors()

	return nil
}

// Close func
func (p *Provider) Close() error {
	p.unscheduleAll()
	return nil
}

func (p *Provider) unscheduleAll() {
	p.scheduler.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, p.GatewayConfig.ID))
}

func (p *Provider) scheduleBridgeSync() error {
	return p.schedule(fmt.Sprintf(scheduleFormatBridge, p.GatewayConfig.ID), p.Config.BridgeSyncInterval, p.updateBridgeDetails)
}

func (p *Provider) scheduleSync() error {
	return p.schedule(fmt.Sprintf(scheduleFormatSync, p.GatewayConfig.ID), p.Config.SyncInterval, p.getUpdate)
}

func (p *Provider) schedule(scheduleID, interval string, triggerFunc func()) error {
	jobSpec := fmt.Sprintf("@every %s", interval)
	err := p.scheduler.AddFunc(scheduleID, jobSpec, triggerFunc)
	if err != nil {
		p.logger.Error("error on adding schedule", zap.Error(err))
		return err
	}
	return nil
}
