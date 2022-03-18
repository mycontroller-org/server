package mysensors

import (
	"fmt"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	scheduleUtils "github.com/mycontroller-org/server/v2/pkg/utils/schedule"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	schedulePrefix       = "mysensors_v2_schedules"
	jobNodeDiscover      = "node_discover"
	autoDiscoverInterval = "15m"
)

// schedules node discover job
func (p *Provider) scheduleNodeDiscover() error {
	scheduleID := scheduleUtils.GetScheduleID(schedulePrefix, p.GatewayConfig.ID, jobNodeDiscover)
	return scheduleUtils.Schedule(scheduleID, fmt.Sprintf("@every %s", autoDiscoverInterval), p.runNodeDiscover)
}

// runs node discover
// this action will not be recorded on server
// this is running in gateway level
func (p *Provider) runNodeDiscover() {
	msg := msgTY.NewMessage(false)
	msg.GatewayID = p.GatewayConfig.ID
	pl := msgTY.NewPayload()
	pl.Key = gatewayTY.ActionDiscoverNodes
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgTY.TypeAction

	// post the message to gateway hardware
	err := p.Post(&msg)
	if err != nil {
		zap.L().Error("error on posting node discover message", zap.String("gatewayId", p.GatewayConfig.ID), zap.Error(err))
	}
}
