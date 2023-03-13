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
	nodeStore  = concurrency.NewStore()
	fwStore    = concurrency.NewStore()
	fwRawStore = concurrency.NewStore()
)

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
func (p *Provider) getFirmwareRaw(id string, fwTypeID, fwVersionID uint16) (*firmwareRaw, error) {
	toFirmwareRaw := func(item interface{}) (*firmwareRaw, error) {
		if fw, ok := item.(*firmwareRaw); ok {
			return fw, nil
		}
		return nil, fmt.Errorf("unknown data received in the place node: %T", item)
	}

	data := fwRawStore.Get(id)
	if data != nil {
		return toFirmwareRaw(data)
	}

	err := p.updateFirmwareFile(id, fwTypeID, fwVersionID)
	if err != nil {
		return nil, err
	}
	data = fwRawStore.Get(id)
	if data != nil {
		return toFirmwareRaw(data)
	}
	return nil, fmt.Errorf("firmware not available. id:%v", id)
}

func (p *Provider) updateFirmwareFile(id string, fwTypeID, fwVersionID uint16) error {
	var hexBytes []byte
	addToStore := func(item interface{}) bool {
		fwBlock, ok := item.(*firmwareTY.FirmwareBlock)
		if !ok {
			p.logger.Error("error on data conversion", zap.String("receivedType", fmt.Sprintf("%T", item)))
			return false
		}
		if hexBytes == nil {
			hexBytes = make([]byte, fwBlock.TotalBytes)
		}
		startPos := int(firmwareTY.BlockSize * fwBlock.BlockNumber)
		for offset, byteData := range fwBlock.Data {
			hexBytes[startPos+offset] = byteData
		}
		if fwBlock.IsFinal {
			receivedCheckSum := fmt.Sprintf("sha256:%x", sha256.Sum256(hexBytes))
			fw, err := p.getFirmware(id)
			if err != nil {
				p.logger.Error("error on getting firmare config", zap.Error(err), zap.String("firmwareId", id))
				return false
			}
			if fw.File.Checksum == receivedCheckSum {
				// convert the hex file to raw format
				fwRaw, err := p.hexByteToLocalFormat(fwTypeID, fwVersionID, hexBytes, firmwareBlockSize)
				if err != nil {
					p.logger.Error("error on converting hex to local format", zap.String("firmwareId", id), zap.Error(err))
					return false
				}
				fwRawStore.Add(id, fwRaw)
			} else {
				p.logger.Info("received firmware checksum mismatch", zap.String("fwID", fw.ID), zap.String("remote", fw.File.Checksum), zap.String("received", receivedCheckSum))
			}
			return false
		}
		return true // continue
	}

	return query.QueryResource(p.logger, p.bus, id, rsTY.TypeFirmware, rsTY.CommandBlocks, nil, addToStore, &firmwareTY.FirmwareBlock{}, queryFirmwareFileTimeout)
}
