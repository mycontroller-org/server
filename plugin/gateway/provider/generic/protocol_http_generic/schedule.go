package http_generic

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

const (
	schedulePrefix      = "generic_provider"
	defaultPoolInterval = "10m"
)

// unschedule all the requests
func (hp *HttpProtocol) unscheduleAll() {
	hp.scheduler.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, hp.GatewayConfig.ID))
}

// schedule a request
func (hp *HttpProtocol) schedule(endpoint string, cfg *HttpConfig) error {
	if cfg.ExecutionInterval == "" {
		cfg.ExecutionInterval = defaultPoolInterval
	}

	triggerFunc := func() {
		rawMsg, err := hp.executeHttpRequest(cfg)
		if err != nil {
			hp.logger.Error("error on executing a request", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("endpoint", endpoint), zap.String("url", cfg.URL), zap.Error(err))
			return
		}
		if rawMsg != nil {
			err = hp.rawMessageHandler(rawMsg)
			if err != nil {
				hp.logger.Error("error on posting a rawmessage", zap.String("gatewayId", hp.GatewayConfig.ID), zap.String("endpoint", endpoint), zap.String("url", cfg.URL), zap.Error(err))
			}
		}
	}

	scheduleID := strings.Join([]string{schedulePrefix, hp.GatewayConfig.ID, endpoint}, "_")
	jobSpec := fmt.Sprintf("@every %s", cfg.ExecutionInterval)
	err := hp.scheduler.AddFunc(scheduleID, jobSpec, triggerFunc)
	if err != nil {
		hp.logger.Error("error on adding schedule", zap.Error(err))
		return err
	}
	return nil
}
