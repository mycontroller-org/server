package mysensors

import (
	"fmt"
	"time"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"go.uber.org/zap"
)

// smartSleepRecentlySeen is the window during which a smart-sleep node is
// considered "awake". Below this threshold, we assume it can receive messages.
const smartSleepRecentlySeen = 30 * time.Second

// shouldRebootForOta reports whether the node opts into automatic reboot for OTA.
// The label must be "true" (case-insensitive). Missing or "false" means no.
func shouldRebootForOta(node *nodeTY.Node) bool {
	if node == nil || node.Labels == nil {
		return false
	}
	return node.Labels.GetBool(LabelBootloaderOtaReboot)
}

// isSleepingNode reports whether the node uses smart sleep.
func isSleepingNode(node *nodeTY.Node) bool {
	if node == nil || node.Labels == nil {
		return false
	}
	return node.Labels.GetBool(LabelSmartSleepNode)
}

// isRecentlySeen reports whether the node was seen within the given window.
func isRecentlySeen(node *nodeTY.Node, threshold time.Duration) bool {
	if node == nil {
		return false
	}
	return time.Since(node.LastSeen) < threshold
}

// maxSleepDuration returns an upper bound on the node's sleep cycle.
// Falls back to 5 minutes when the label is missing.
func maxSleepDuration(node *nodeTY.Node) time.Duration {
	if node == nil || node.Labels == nil {
		return 5 * time.Minute
	}
	ms := int64(node.Labels.GetInt(FieldSleepDuration))
	if ms <= 0 {
		return 5 * time.Minute
	}
	d := time.Duration(ms) * time.Millisecond
	if d < time.Minute {
		d = time.Minute
	}
	if d > time.Hour {
		d = time.Hour
	}
	return d
}

// maybeSendRebootForOta is called before serving a firmware-update action.
// It sends I_REBOOT to awake nodes, or queues it for sleeping ones.
//
// A reboot is needed when either:
//   - the user explicitly opted in with bootloader_ota_reboot=true, or
//   - the detected bootloader is a MYSBootloader variant (any 1.x).
//
// Protocol 3.1 / DualOptiboot nodes do not need a reboot: they enter OTA on
// their own when they receive the firmware config request.
func (p *Provider) maybeSendRebootForOta(msg *msgTY.Message) error {
	node, err := p.getNode(msg.GatewayID, msg.NodeID)
	if err != nil {
		return fmt.Errorf("load node for OTA reboot check: %w", err)
	}

	explicitOptIn := shouldRebootForOta(node)
	detectedMYSBootloader := isMYSBootloaderKind(bootloaderKindForNode(node))

	if !explicitOptIn && !detectedMYSBootloader {
		p.logger.Debug("no pre-OTA reboot needed",
			zap.String("nodeId", node.NodeID),
			zap.String("gatewayId", node.GatewayID),
			zap.String("bootloaderKind", bootloaderKindForNode(node)),
			zap.Bool("explicitOptIn", explicitOptIn),
		)
		return nil
	}

	// A reboot is needed. Queue it if the node is a smart-sleep node that is
	// currently asleep; otherwise send it immediately.
	if isSleepingNode(node) && !isRecentlySeen(node, smartSleepRecentlySeen) {
		p.queueRebootForNode(node)
		p.logger.Info("node is sleeping, I_REBOOT queued until next wake",
			zap.String("nodeId", node.NodeID),
			zap.String("gatewayId", node.GatewayID),
			zap.String("bootloaderKind", bootloaderKindForNode(node)),
			zap.Bool("explicitOptIn", explicitOptIn),
			zap.String("sleepDuration", node.Labels.Get(FieldSleepDuration)),
			zap.Duration("maxWait", maxSleepDuration(node)),
		)
		return nil
	}

	return p.sendRebootNow(node)
}

// sendRebootNow sends an I_REBOOT to the node immediately.
func (p *Provider) sendRebootNow(node *nodeTY.Node) error {
	rebootMsg := msgTY.NewMessage(false)
	rebootMsg.GatewayID = node.GatewayID
	rebootMsg.NodeID = node.NodeID
	rebootMsg.Type = msgTY.TypeAction

	pl := msgTY.NewPayload()
	pl.Key = nodeTY.ActionReboot
	rebootMsg.Payloads = []msgTY.Payload{pl}

	if err := p.Post(&rebootMsg); err != nil {
		return fmt.Errorf("post reboot for OTA: %w", err)
	}

	p.logger.Info("sent I_REBOOT before OTA",
		zap.String("nodeId", node.NodeID),
		zap.String("gatewayId", node.GatewayID),
		zap.String("bootloaderKind", node.Labels.Get(LabelBootloaderKind)),
		zap.String("bootloaderVersion", node.Labels.Get(LabelBootloaderVersion)),
	)
	return nil
}

// queueRebootForNode stores a pending I_REBOOT for delivery when the node wakes.
func (p *Provider) queueRebootForNode(node *nodeTY.Node) {
	if node == nil {
		return
	}
	key := p.getNodeStoreID(node.GatewayID, node.NodeID)
	otaPendingRebootStore.Add(key, time.Now())
}

// deliverQueuedRebootIfAny sends a pending I_REBOOT when the node is known to be awake.
// Safe to call multiple times; the entry is removed on first successful delivery.
func (p *Provider) deliverQueuedRebootIfAny(gatewayID, nodeID string) {
	key := p.getNodeStoreID(gatewayID, nodeID)
	v := otaPendingRebootStore.Get(key)
	if v == nil {
		return
	}
	otaPendingRebootStore.Remove(key)

	node, err := p.getNode(gatewayID, nodeID)
	if err != nil {
		p.logger.Error("failed to load node for queued reboot delivery",
			zap.String("gatewayId", gatewayID),
			zap.String("nodeId", nodeID),
			zap.Error(err),
		)
		return
	}

	p.logger.Info("delivering queued I_REBOOT to awake node",
		zap.String("nodeId", node.NodeID),
		zap.String("gatewayId", node.GatewayID),
	)

	// Allow a short delay so the node has completed its wake-up sequence
	// (heartbeat, I_CONFIG, presentation, ...) before receiving I_REBOOT.
	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := p.sendRebootNow(node); err != nil {
			p.logger.Error("failed to deliver queued reboot",
				zap.String("nodeId", node.NodeID),
				zap.Error(err),
			)
		}
	}()
}

// pendingRebootCount returns the number of queued reboots (used by tests).
func pendingRebootCount() int {
	return len(otaPendingRebootStore.Keys())
}
