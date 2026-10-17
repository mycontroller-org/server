package mysensors

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"go.uber.org/zap"
)

// sleepRecentlySeen is the same threshold as the OTA reboot logic.
// A node seen within this window is considered awake.
const sleepRecentlySeen = 30 * time.Second

// sleepCommandVPercentage is the numeric code for V_PERCENTAGE
// (see setReqFieldMapForRx in constant.go).
const sleepCommandVPercentage = "3"

// sleepCommandMaxAge is the TTL of a queued command.
// Beyond this, the command is considered stale and purged.
const sleepCommandMaxAge = 2 * time.Hour

// parseSleepDurationPayload extracts the optional sensor_id and the percentage.
//
// Accepted formats:
//
//	"32"         -> ("", 32, nil)
//	"206:32"     -> ("206", 32, nil)
//	" 206 : 32 " -> ("206", 32, nil)
//	"foo"        -> ("", 0, err)
//	"206:200"    -> ("", 0, err)
//	"206:foo"    -> ("", 0, err)
func parseSleepDurationPayload(raw string) (sensorID string, percent uint8, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, fmt.Errorf("empty payload")
	}
	valueStr := raw
	if idx := strings.Index(raw, ":"); idx >= 0 {
		sensorID = strings.TrimSpace(raw[:idx])
		valueStr = strings.TrimSpace(raw[idx+1:])
	}
	pct, convErr := strconv.Atoi(valueStr)
	if convErr != nil || pct < 0 || pct > 100 {
		return "", 0, fmt.Errorf("invalid sleep percentage: %q (expected integer 0-100)", valueStr)
	}
	return sensorID, uint8(pct), nil
}

// sendSleepCommand sends (or queues) a sleep-duration change to a smart_sleep
// node.
//
// If the node is awake (recent contact), the command goes out immediately.
// Otherwise it is stored and delivered on the next I_PRE_SLEEP_NOTIFICATION
// or I_HEARTBEAT_RESPONSE (see deliverQueuedSleepCommandIfAny).
func (p *Provider) sendSleepCommand(node *nodeTY.Node, sensorID string, percent uint8, requestID string) error {
	if node == nil {
		return fmt.Errorf("nil node")
	}
	if sensorID == "" {
		return fmt.Errorf("sleep command sensor id is empty")
	}
	if percent > 100 {
		return fmt.Errorf("invalid sleep percentage: %d", percent)
	}
	if !isSleepingNode(node) {
		return fmt.Errorf("node %s is not a smart_sleep node", node.ID)
	}

	if isRecentlySeen(node, sleepRecentlySeen) {
		return p.sendSleepCommandNow(node, sensorID, percent)
	}

	p.queueSleepCommand(node, sensorID, percent, requestID)
	p.logger.Info("sleep command queued until next wake",
		zap.String("nodeId", node.NodeID),
		zap.String("gatewayId", node.GatewayID),
		zap.String("sensorId", sensorID),
		zap.Uint8("percent", percent),
		zap.String("requestId", requestID),
		zap.String("sleepDuration", node.Labels.Get(FieldSleepDuration)),
		zap.Duration("maxWait", maxSleepDuration(node)),
	)
	return nil
}

// sendSleepCommandNow sends the C_SET/V_PERCENTAGE to the node immediately.
func (p *Provider) sendSleepCommandNow(node *nodeTY.Node, sensorID string, percent uint8) error {
	sleepMsg := msgTY.NewMessage(false)
	sleepMsg.GatewayID = node.GatewayID
	sleepMsg.NodeID = node.NodeID
	sleepMsg.SourceID = sensorID
	sleepMsg.Type = msgTY.TypeSet

	pl := msgTY.NewPayload()
	pl.Key = "V_PERCENTAGE"
	pl.SetValue(fmt.Sprintf("%d", percent))
	pl.Labels.Set(LabelType, sleepCommandVPercentage)

	sleepMsg.Payloads = []msgTY.Payload{pl}

	if err := p.Post(&sleepMsg); err != nil {
		return fmt.Errorf("post sleep command: %w", err)
	}

	p.logger.Info("sent sleep command",
		zap.String("nodeId", node.NodeID),
		zap.String("gatewayId", node.GatewayID),
		zap.String("sensorId", sensorID),
		zap.Uint8("percent", percent),
	)
	return nil
}

// queueSleepCommand stores the command for later delivery.
func (p *Provider) queueSleepCommand(node *nodeTY.Node, sensorID string, percent uint8, requestID string) {
	if node == nil {
		return
	}
	key := p.getNodeStoreID(node.GatewayID, node.NodeID)
	pendingSleepStore.Add(key, pendingSleepCommand{
		SensorID:  sensorID,
		Percent:   percent,
		QueuedAt:  time.Now(),
		RequestID: requestID,
	})
}

// deliverQueuedSleepCommandIfAny sends the queued command when the node wakes.
// Called from updateNodeInternalMessages() on I_PRE_SLEEP_NOTIFICATION /
// I_HEARTBEAT_RESPONSE, same as the OTA reboot delivery.
//
// Idempotent: the entry is removed before sending to avoid duplicates.
func (p *Provider) deliverQueuedSleepCommandIfAny(gatewayID, nodeID string) {
	key := p.getNodeStoreID(gatewayID, nodeID)
	v := pendingSleepStore.Get(key)
	if v == nil {
		return
	}
	cmd, ok := v.(pendingSleepCommand)
	if !ok {
		pendingSleepStore.Remove(key)
		return
	}
	if time.Since(cmd.QueuedAt) > sleepCommandMaxAge {
		pendingSleepStore.Remove(key)
		p.logger.Warn("pending sleep command expired",
			zap.String("gatewayId", gatewayID),
			zap.String("nodeId", nodeID),
		)
		return
	}
	pendingSleepStore.Remove(key)

	node, err := p.getNode(gatewayID, nodeID)
	if err != nil {
		p.logger.Error("failed to load node for queued sleep command",
			zap.String("gatewayId", gatewayID),
			zap.String("nodeId", nodeID),
			zap.Error(err),
		)
		return
	}

	p.logger.Info("delivering queued sleep command to awake node",
		zap.String("nodeId", node.NodeID),
		zap.String("gatewayId", node.GatewayID),
		zap.String("sensorId", cmd.SensorID),
		zap.Uint8("percent", cmd.Percent),
	)

	// Allow a short delay so the node has completed its wake-up sequence
	// (heartbeat, I_CONFIG, presentation, ...) before receiving the C_SET.
	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := p.sendSleepCommandNow(node, cmd.SensorID, cmd.Percent); err != nil {
			p.logger.Error("failed to deliver queued sleep command",
				zap.String("nodeId", node.NodeID),
				zap.Error(err),
			)
		}
	}()
}

// pendingSleepCount is used by tests.
func pendingSleepCount() int {
	return len(pendingSleepStore.Keys())
}

// purgeExpiredSleepCommands is scheduled like firmwareRawPurge().
// Removes commands that are too old.
func purgeExpiredSleepCommands() {
	now := time.Now()
	for _, key := range pendingSleepStore.Keys() {
		v := pendingSleepStore.Get(key)
		cmd, ok := v.(pendingSleepCommand)
		if !ok || now.Sub(cmd.QueuedAt) > sleepCommandMaxAge {
			pendingSleepStore.Remove(key)
		}
	}
}
