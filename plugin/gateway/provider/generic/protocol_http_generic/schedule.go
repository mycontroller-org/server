package http_generic

import (
	"fmt"

	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"go.uber.org/zap"
)

const (
	schedulePrefix      = "generic_provider"
	defaultPoolInterval = "10m"
)

// unschedule all the requests
func (hp *HttpProtocol) unscheduleAll() {
	coreScheduler.SVC.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, hp.GatewayConfig.ID))
}

// schedule a request
func (hp *HttpProtocol) schedule(endpoint string, cfg *HttpConfig) error {
	if cfg.ExecutionInterval == "" {
		cfg.ExecutionInterval = defaultPoolInterval
	}

	triggerFunc := func() {
		rawMsg, err := hp.executeHttpRequest(cfg)
		if err != nil {
			zap.L().Error("error on executing a request", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("endpoint", endpoint), zap.String("url", cfg.URL), zap.Error(err))
			return
		}
		if rawMsg != nil {
			err = hp.rawMessageHandler(rawMsg)
			if err != nil {
				zap.L().Error("error on posting a rawmessage", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("endpoint", endpoint), zap.String("url", cfg.URL), zap.Error(err))
			}
		}
	}

	scheduleID := fmt.Sprintf("%s_%s_%s", schedulePrefix, hp.GatewayConfig.ID, endpoint)
	cronSpec := fmt.Sprintf("@every %s", cfg.ExecutionInterval)
	err := coreScheduler.SVC.AddFunc(scheduleID, cronSpec, triggerFunc)
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
		return err
	}
	zap.L().Debug("added a schedule", zap.String("schedulerID", scheduleID), zap.String("interval", cfg.ExecutionInterval))
	return nil
}
