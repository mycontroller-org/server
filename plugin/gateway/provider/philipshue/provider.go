package philipshue

import (
	"fmt"
	"time"

	"github.com/amimof/huego"
	gwML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	coreScheduler "github.com/mycontroller-org/backend/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	schedulePrefix            = "philipshue_status"
	scheduleFormatSync        = "%s_sync"
	scheduleFormatBridge      = "%s_bridge"
	defaultSyncInterval       = "30m"
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
	Config        Config
	GatewayConfig *gwML.Config
	bridge        *huego.Bridge
}

// Init provider
func Init(gatewayCfg *gwML.Config) (*Provider, error) {
	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameNone, gatewayCfg.Provider, &cfg)
	if err != nil {
		return nil, err
	}

	_, err = time.ParseDuration(cfg.SyncInterval)
	if err != nil {
		zap.L().Debug("invalid sync interval supplied. set to default", zap.String("input", cfg.SyncInterval), zap.String("default", defaultSyncInterval))
		cfg.SyncInterval = defaultSyncInterval
	}

	_, err = time.ParseDuration(cfg.BridgeSyncInterval)
	if err != nil {
		zap.L().Debug("invalid bridge sync interval supplied. set to default", zap.String("input", cfg.BridgeSyncInterval), zap.String("default", defaultBridgeSyncInterval))
		cfg.BridgeSyncInterval = defaultBridgeSyncInterval
	}

	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayCfg,
	}
	zap.L().Debug("Config details", zap.Any("received", gatewayCfg.Provider), zap.Any("converted", cfg))
	return provider, nil
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgML.RawMessage) error) error {
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
	// do internal works
	p.unscheduleAll()
	return nil
}

func (p *Provider) unscheduleAll() {
	coreScheduler.SVC.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, p.GatewayConfig.ID))
}

func (p *Provider) scheduleBridgeSync() error {
	return p.schedule(fmt.Sprintf(scheduleFormatBridge, p.GatewayConfig.ID), p.Config.BridgeSyncInterval, p.updateBridgeDetails)
}

func (p *Provider) scheduleSync() error {
	return p.schedule(fmt.Sprintf(scheduleFormatSync, p.GatewayConfig.ID), p.Config.SyncInterval, p.getUpdate)
}

func (p *Provider) schedule(schedulerID, interval string, triggerFunc func()) error {
	cronSpec := fmt.Sprintf("@every %s", interval)
	err := coreScheduler.SVC.AddFunc(schedulerID, cronSpec, triggerFunc)
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
		return err
	}
	zap.L().Debug("added a schedule", zap.String("schedulerID", schedulerID), zap.String("interval", interval))
	return nil
}
