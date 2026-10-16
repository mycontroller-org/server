package mysensors

import (
	"encoding/binary"
	hexENC "encoding/hex"
	"fmt"
	"strings"

	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"go.uber.org/zap"
)

// Labels set on nodes to remember the detected bootloader.
const (
	LabelBootloaderKind    = "bootloader_kind"    // e.g. "mysbootloader-1.4", "protocol-3.1"
	LabelBootloaderVersion = "bootloader_version" // e.g. "1.4" (MYSBootloader only)
)

// Bootloader kinds recognised by the controller.
const (
	BootloaderKindMYSBootloader10 = "mysbootloader-1.0"
	BootloaderKindMYSBootloader13 = "mysbootloader-1.3"
	BootloaderKindMYSBootloader14 = "mysbootloader-1.4"
	BootloaderKindProtocol31      = "protocol-3.1"
	BootloaderKindUnknown         = "unknown"
)

// isMYSBootloaderKind reports whether a kind string designates any MYSBootloader.
func isMYSBootloaderKind(kind string) bool {
	return strings.HasPrefix(kind, "mysbootloader")
}

// detectBootloaderKind classifies the bootloader from a ST_FIRMWARE_CONFIG_REQUEST
// payload. Uses payload length and (for the 10-byte format) the MYSBootloader
// BLVersion field.
//
// Signatures:
//
//	6  bytes : MYSBootloader < 1.3 (type, version, blocks)
//	10 bytes : MYSBootloader >= 1.3 (type, version, blocks, crc, BLVersion)
//	11 bytes : protocol 3.1 (10 bytes + blockSize)
//	12+ bytes: protocol 3.1 extended (custom)
func detectBootloaderKind(requestHex string) string {
	b, err := hexENC.DecodeString(sanitizeOtaHex(requestHex))
	if err != nil || len(b) == 0 {
		return ""
	}
	switch len(b) {
	case 6:
		return BootloaderKindMYSBootloader10
	case 10:
		// BLVersion is a little-endian uint16 at offset 8.
		// MYSBootloader packs (minor << 8) | major.
		blv := binary.LittleEndian.Uint16(b[8:10])
		major := blv & 0xFF
		minor := (blv >> 8) & 0xFF
		if major == 1 {
			switch minor {
			case 3:
				return BootloaderKindMYSBootloader13
			case 4:
				return BootloaderKindMYSBootloader14
			default:
				if minor >= 3 && minor <= 9 {
					return fmt.Sprintf("mysbootloader-1.%d", minor)
				}
				return "mysbootloader-compatible"
			}
		}
		return BootloaderKindUnknown
	case 11, 12, 18:
		// Protocol 3.1: at offset 10, blockSize (must be a positive multiple of 8).
		if b[10] >= 8 && b[10]%8 == 0 {
			return BootloaderKindProtocol31
		}
		return BootloaderKindUnknown
	default:
		if len(b) >= 10 && b[10] >= 8 && b[10]%8 == 0 {
			return BootloaderKindProtocol31
		}
		return BootloaderKindUnknown
	}
}

// extractBootloaderVersion returns a human string like "1.4" for MYSBootloader,
// or "" if the version cannot be determined.
func extractBootloaderVersion(requestHex string) string {
	b, err := hexENC.DecodeString(sanitizeOtaHex(requestHex))
	if err != nil || len(b) < 10 {
		return ""
	}
	if len(b) != 10 && len(b) != 11 && len(b) != 12 && len(b) != 18 {
		return ""
	}
	blv := binary.LittleEndian.Uint16(b[8:10])
	major := blv & 0xFF
	minor := (blv >> 8) & 0xFF
	if major == 0 && minor == 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d", major, minor)
}

// rememberBootloaderKind persists the detected bootloader in the node labels.
// Does nothing if nothing changed, to avoid unnecessary writes.
func (p *Provider) rememberBootloaderKind(node *nodeTY.Node, requestHex string) {
	if node == nil || node.Labels == nil || requestHex == "" {
		return
	}
	kind := detectBootloaderKind(requestHex)
	if kind == "" {
		return
	}
	version := extractBootloaderVersion(requestHex)

	changed := false
	if node.Labels.Get(LabelBootloaderKind) != kind {
		node.Labels.Set(LabelBootloaderKind, kind)
		changed = true
	}
	if version != "" && node.Labels.Get(LabelBootloaderVersion) != version {
		node.Labels.Set(LabelBootloaderVersion, version)
		changed = true
	}
	if !changed {
		return
	}
	p.setNodeLabels(node)
	p.logger.Info("detected bootloader",
		zap.String("nodeId", node.NodeID),
		zap.String("gatewayId", node.GatewayID),
		zap.String("kind", kind),
		zap.String("version", version),
		zap.Int("payloadBytes", len(requestHex)/2),
	)
}

// bootloaderKindForNode returns the persisted bootloader kind for a node, or "".
func bootloaderKindForNode(node *nodeTY.Node) string {
	if node == nil || node.Labels == nil {
		return ""
	}
	return node.Labels.Get(LabelBootloaderKind)
}

// IsMYSBootloader reports whether the node's persisted bootloader kind
// designates any MYSBootloader variant.
func (p *Provider) IsMYSBootloader(node *nodeTY.Node) bool {
	return isMYSBootloaderKind(bootloaderKindForNode(node))
}

// BootloaderKind returns the persisted bootloader kind for a node, or "".
func (p *Provider) BootloaderKind(node *nodeTY.Node) string {
	return bootloaderKindForNode(node)
}
