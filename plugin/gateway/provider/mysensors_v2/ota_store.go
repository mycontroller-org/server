package mysensors

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/utils/bus_utils/query"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	"go.uber.org/zap"
)

var (
	nodeStore        = concurrency.NewStore()
	fwStore          = concurrency.NewStore()
	fwRawStore       = concurrency.NewStore()
	fotaScriptStore  = concurrency.NewStore() // data_repository id → *fotaScriptBundle
	fotaSessionStore = concurrency.NewStore() // gateway_node → *fotaSession
)

// fotaSession is the firmware + slice size advertised in onConfig, reused for every block.
type fotaSession struct {
	FirmwareID string
	BlockSize  int
}

func firmwareRawPurge() {
	for _, fwID := range fwRawStore.Keys() {
		fwInf := fwRawStore.Get(fwID)
		fw, ok := fwInf.(*firmwareRaw)
		if !ok {
			continue
		}
		if time.Since(fw.LastAccess) >= firmwarePurgeInactiveTime { // eligible for purging
			fwRawStore.Remove(fwID)
		}
	}
}

// getNode returns the node
func (p *Provider) getNode(gatewayID, nodeID string) (*nodeTY.Node, error) {
	id := p.getNodeStoreID(gatewayID, nodeID)

	toNode := func(item interface{}) (*nodeTY.Node, error) {
		if node, ok := item.(*nodeTY.Node); ok {
			return node, nil
		}
		return nil, fmt.Errorf("unknown data received in the place node: %T", item)
	}

	data := nodeStore.Get(id)
	if data != nil {
		return toNode(data)
	}

	err := p.updateNode(gatewayID, nodeID)
	if err != nil {
		return nil, err
	}
	data = nodeStore.Get(id)
	if data != nil {
		return toNode(data)
	}
	return nil, fmt.Errorf("node not available. gatewayID:%s, nodeID:%s", gatewayID, nodeID)
}

// getNode returns the node
func (p *Provider) getFirmware(id string) (*firmwareTY.Firmware, error) {
	toFirmware := func(item interface{}) (*firmwareTY.Firmware, error) {
		if fw, ok := item.(*firmwareTY.Firmware); ok {
			return fw, nil
		}
		return nil, fmt.Errorf("unknown data received in the place node: %T", item)
	}

	data := fwStore.Get(id)
	if data != nil {
		return toFirmware(data)
	}

	err := p.updateFirmware(id)
	if err != nil {
		return nil, err
	}
	data = fwStore.Get(id)
	if data != nil {
		return toFirmware(data)
	}
	return nil, fmt.Errorf("firmware not available. id:%v", id)
}

func (p *Provider) getNodeStoreID(gatewayID, nodeID string) string {
	return fmt.Sprintf("%s_%s", gatewayID, nodeID)
}

func (p *Provider) updateNode(gatewayID, nodeID string) error {
	ids := map[string]interface{}{
		types.KeyGatewayID: gatewayID,
		types.KeyNodeID:    nodeID,
	}

	addToStore := func(item interface{}) bool {
		node, ok := item.(*nodeTY.Node)
		if !ok {
			p.logger.Error("error on data conversion", zap.String("receivedType", fmt.Sprintf("%T", item)))
			return false
		}
		nodeStore.Add(p.getNodeStoreID(node.GatewayID, node.NodeID), node)
		return false
	}
	return query.QueryResource(p.logger, p.bus, "", rsTY.TypeNode, rsTY.CommandGet, ids, addToStore, &nodeTY.Node{}, queryTimeout)
}

func (p *Provider) updateFirmware(id string) error {
	addToStore := func(item interface{}) bool {
		firmware, ok := item.(*firmwareTY.Firmware)
		if !ok {
			p.logger.Error("error on data conversion", zap.String("receivedType", fmt.Sprintf("%T", item)))
			return false
		}
		fwStore.Add(firmware.ID, firmware)
		return false
	}
	return query.QueryResource(p.logger, p.bus, id, rsTY.TypeFirmware, rsTY.CommandGet, nil, addToStore, &firmwareTY.Firmware{}, queryTimeout)
}

// getFirmwareRaw func
func (p *Provider) getFirmwareRaw(id string, fwTypeID, fwVersionID uint16, blockSize int) (*firmwareRaw, error) {
	toFirmwareRaw := func(item interface{}) (*firmwareRaw, error) {
		if fw, ok := item.(*firmwareRaw); ok {
			return fw, nil
		}
		return nil, fmt.Errorf("unknown data received in the place node: %T", item)
	}

	if blockSize <= 0 {
		blockSize = defaultFirmwareBlockSize
	}
	cacheKey := firmwareRawCacheKey(id, blockSize)
	data := fwRawStore.Get(cacheKey)
	if data != nil {
		return toFirmwareRaw(data)
	}

	err := p.updateFirmwareFile(id, fwTypeID, fwVersionID, blockSize)
	if err != nil {
		return nil, err
	}
	data = fwRawStore.Get(cacheKey)
	if data != nil {
		return toFirmwareRaw(data)
	}
	return nil, fmt.Errorf("firmware not available. id:%v", id)
}

func assembleFirmwareBytes(blocks map[int][]byte, totalBytes int) ([]byte, bool) {
	if totalBytes <= 0 || len(blocks) == 0 {
		return nil, false
	}
	out := make([]byte, totalBytes)
	seen := make([]bool, totalBytes)
	for n, data := range blocks {
		start := firmwareTY.BlockSize * n
		for i, v := range data {
			pos := start + i
			if pos >= totalBytes {
				break
			}
			out[pos] = v
			seen[pos] = true
		}
	}
	for i := 0; i < totalBytes; i++ {
		if !seen[i] {
			return nil, false
		}
	}
	return out, true
}

func (p *Provider) updateFirmwareFile(id string, fwTypeID, fwVersionID uint16, blockSize int) error {
	// Load metadata first. Do not CommandGet from inside the block callback;
	// that nested query races with remaining block replies on the same bus.
	fw, err := p.getFirmware(id)
	if err != nil {
		return err
	}

	blocks := map[int][]byte{}
	totalBytes := 0
	addToStore := func(item interface{}) bool {
		fwBlock, ok := item.(*firmwareTY.FirmwareBlock)
		if !ok {
			p.logger.Error("error on data conversion", zap.String("receivedType", fmt.Sprintf("%T", item)))
			return false
		}
		if fwBlock.TotalBytes > 0 {
			totalBytes = fwBlock.TotalBytes
		}
		cp := make([]byte, len(fwBlock.Data))
		copy(cp, fwBlock.Data)
		blocks[fwBlock.BlockNumber] = cp

		hexBytes, complete := assembleFirmwareBytes(blocks, totalBytes)
		if !complete {
			return true
		}
		receivedCheckSum := fmt.Sprintf("sha256:%x", sha256.Sum256(hexBytes))
		if fw.File.Checksum != receivedCheckSum {
			p.logger.Info("firmware file checksum mismatch (re-upload the firmware in the UI)",
				zap.String("fwID", fw.ID),
				zap.String("file", fw.File.Name),
				zap.Int("bytes", len(hexBytes)),
				zap.Int("blockCount", len(blocks)),
				zap.String("stored", fw.File.Checksum),
				zap.String("computed", receivedCheckSum),
			)
			return false
		}
		fwRaw, err := p.hexByteToLocalFormat(fwTypeID, fwVersionID, hexBytes, blockSize)
		if err != nil {
			p.logger.Error("error on converting hex to local format", zap.String("firmwareId", id), zap.Error(err))
			return false
		}
		fwRawStore.Add(firmwareRawCacheKey(id, blockSize), fwRaw)
		return false
	}

	return query.QueryResource(p.logger, p.bus, id, rsTY.TypeFirmware, rsTY.CommandBlocks, nil, addToStore, &firmwareTY.FirmwareBlock{}, queryFirmwareFileTimeout)
}
