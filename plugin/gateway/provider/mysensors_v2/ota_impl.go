package mysensors

import (
	"bytes"
	"encoding/binary"
	hexENC "encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"go.uber.org/zap"
)

// executeFirmwareConfigRequest executes firmware config request and response with hex payload
func (p *Provider) executeFirmwareConfigRequest(msg *msgTY.Message) (string, error) {
	startTime := time.Now()
	rxPL := msg.Payloads[0].Value.String()

	// convert the received hex to matching struct format
	fwCfgReq := &firmwareConfigRequest{}
	if rxPL != "" {
		err := toStruct(rxPL, fwCfgReq)
		if err != nil {
			p.logger.Error("error on converting firmwareConfigRequest", zap.String("payload", rxPL), zap.Error(err))
			return "", err
		}
	}

	node, err := p.getNode(msg.GatewayID, msg.NodeID)
	if err != nil {
		p.logger.Error("error to get node details", zap.Any("msg", msg), zap.Error(err))
		return "", err
	}

	// get firmware raw format
	fwRaw, err := p.fetchFirmware(node, fwCfgReq.Type, fwCfgReq.Version, false)
	if err != nil {
		p.logger.Error("error to get firmware", zap.Any("fwCfgReq", fwCfgReq), zap.Error(err))
		return "", err
	}
	fwRaw.LastAccess = time.Now()

	// create firmware config response struct and update required values
	fwCfgRes := &firmwareConfigResponse{}

	// if erase eeprom set for this node, update erase eeprom command and clear the label on the node detail
	if node.Labels.GetBool(LabelEraseEEPROM) {
		p.logger.Debug("erase EEPROM enabled, sending erase EEPROM command to the node", zap.String("nodeId", node.ID))
		// set erase command
		fwCfgRes.SetEraseEEPROM()
		// remove erase config data from node
		node.Labels.Set(LabelEraseEEPROM, "false")

		p.setNodeLabels(node)
	} else { // update assigned firmware config details
		fwCfgRes.Type = fwRaw.Type
		fwCfgRes.Version = fwRaw.Version
		fwCfgRes.Blocks = fwRaw.Blocks
		fwCfgRes.CRC = fwRaw.CRC
	}
	p.logger.Debug("sending a firmware config respose", zap.Any("request", fwCfgReq), zap.Any("response", fwCfgRes), zap.String("timeTaken", time.Since(startTime).String()))

	// convert the struct to hex string and return
	return p.toHex(fwCfgRes)
}

// executeFirmwareRequest executes firmware request and response with hex payload
func (p *Provider) executeFirmwareRequest(msg *msgTY.Message) (string, error) {
	rxPL := msg.Payloads[0].Value.String()
	startTime := time.Now()

	// convert the received hex to matching struct format
	fwReq := &firmwareRequest{}
	err := toStruct(rxPL, fwReq)
	if err != nil {
		p.logger.Error("error on converting firmwareRequest", zap.String("payload", rxPL), zap.Error(err))
		return "", err
	}

	node, err := p.getNode(msg.GatewayID, msg.NodeID)
	if err != nil {
		p.logger.Error("error to get node details", zap.Any("msg", msg), zap.Error(err))
		return "", err
	}

	// get firmware raw format
	fwRaw, err := p.fetchFirmware(node, fwReq.Type, fwReq.Version, true)
	if err != nil {
		return "", fmt.Errorf("error on getting firmware. %s", err.Error())
	}
	fwRaw.LastAccess = time.Now()

	// create firmware config response struct and update required values
	fwRes := &firmwareResponse{
		Type:    fwReq.Type,
		Version: fwReq.Version,
		Block:   fwReq.Block,
	}

	startAddr := fwReq.Block * firmwareBlockSize
	endAddr := startAddr + firmwareBlockSize
	copy(fwRes.Data[:], fwRaw.Data[startAddr:endAddr])
	p.logger.Debug("sending a firmware respose", zap.Any("request", fwReq), zap.Any("response", fwRes), zap.String("timeTaken", time.Since(startTime).String()))

	p.updateFirmwareProgressStatus(node, int(fwReq.Block), len(fwRaw.Data))

	// convert the struct to hex string and return
	return p.toHex(fwRes)
}

// fetchFirmware looks requested firmware on memory store,
// if not available, loads it from disk
func (p *Provider) fetchFirmware(node *nodeTY.Node, typeID, versionID uint16, verifyID bool) (*firmwareRaw, error) {
	// get mapped firmware by id
	fwID := node.Labels.Get(types.LabelNodeAssignedFirmware)
	if fwID == "" {
		return nil, fmt.Errorf("firmware not assigned for this node. gatewayId:%s, nodeId:%s, typeId:%d, versionId:%d", node.GatewayID, node.NodeID, typeID, versionID)
	}

	// lambda function to load firmware
	loadFirmwareRawFn := func() (*firmwareRaw, error) {

		fw, err := p.getFirmware(fwID)
		if err != nil {
			p.logger.Error("error to get firmware raw", zap.Any("fwID", fwID), zap.Error(err))
			return nil, err
		}

		// get mysensor specific ids
		if fw.Labels.Get(LabelFirmwareTypeID) == "" || fw.Labels.Get(LabelFirmwareVersionID) == "" {
			return nil, fmt.Errorf("firmware '%s' or '%s' labels are not set", LabelFirmwareTypeID, LabelFirmwareVersionID)
		}
		fwTypeID := uint16(fw.Labels.GetInt(LabelFirmwareTypeID))
		fwVersionID := uint16(fw.Labels.GetInt(LabelFirmwareVersionID))

		fwRaw, err := p.getFirmwareRaw(fw.ID, fwTypeID, fwVersionID)
		if err != nil {
			p.logger.Error("error on getting firmware data", zap.String("firmwareId", fw.ID), zap.Error(err))
			return nil, err
		}

		// keep it on memory store
		fwRawStore.Add(fwID, fwRaw)
		return fwRaw, nil
	}

	// check firmware on memory store
	// if not found, load it from disk
	var fwRaw *firmwareRaw
	fwRawInf := fwRawStore.Get(fwID)
	if fwRawInf == nil {
		_fwRaw, err := loadFirmwareRawFn()
		if err != nil {
			return nil, err
		}
		fwRaw = _fwRaw
	} else {
		_fwRaw, ok := fwRawInf.(*firmwareRaw)
		if !ok {
			return nil, fmt.Errorf("error on converting target type. firmwareID: %s", fwID)
		}
		fwRaw = _fwRaw
	}

	if verifyID { // verify firmware ids
		if fwRaw.Type != typeID || fwRaw.Version != versionID {
			return nil, fmt.Errorf("requested firmware type id or version id not matching[Req, Avl], TypeId:[%v, %v], VersionId:[%v, %v]",
				typeID, fwRaw.Type, versionID, fwRaw.Version)
		}
	}

	return fwRaw, nil
}

// Source: https://en.wikipedia.org/wiki/Intel_HEX
// https://github.com/mycontroller-org/mycontroller-v1-legacy/blob/1.5.0.Final/modules/core/src/main/java/org/mycontroller/standalone/firmware/FirmwareUtils.java#L118
// https://github.com/mysensors/MySensorsSampleController/blob/9dbae76081a9c080d5fdd68fba9870626025343f/NodeJsController.js#L172
// I8HEX files use only record types 00 and 01 (16-bit addresses)
// 00 - data, 01 - End
// Example,
//
//	:10010000214601360121470136007EFE09D2190140
//	:100110002146017E17C20001FF5F16002148011928
//	:10012000194E79234623965778239EDA3F01B2CAA7
//	:100130003F0156702B5E712B722B732146013421C7
//	:00000001FF
//	:(start) xx(byte count) xxxx(address) xx(record type) xxx...xx(data, checksum)
func (p *Provider) hexByteToLocalFormat(typeID, versionID uint16, hexByte []byte, blockSize int) (*firmwareRaw, error) {
	hexString := string(hexByte)
	hexString = strings.ReplaceAll(hexString, "\r", "") // remove all "\r" char
	hexLines := strings.Split(hexString, "\n")          // split as separate lines

	actualData := make([]byte, 0)
	for index, line := range hexLines {
		line = strings.TrimSpace(line) // remove spaces if any
		if len(line) == 0 {
			continue
		}
		// first char of the line should be ':'
		if line[0] != ':' {
			return nil, fmt.Errorf("hex line not started with the char ':', line number:%d, data:%s", index+1, line)
		}

		// we are not going to use byte count, address and checksum
		// ignore those fields
		// byte count => line[1:3]
		// address => line[3:7]

		recordType, err := strconv.ParseInt(line[7:9], 16, 64)
		if err != nil {
			return nil, err
		}

		if recordType != 0 {
			continue
		}

		// get only data bytes and convert to bytes from string bytes
		data := line[9 : len(line)-2]
		dataBytes, err := hexENC.DecodeString(data)
		if err != nil {
			p.logger.Error("failed", zap.Any("data", data), zap.Error(err))
			return nil, err
		}
		// include it to our main slice
		actualData = append(actualData, dataBytes...)
	}
	// check the processed bytes length
	if len(actualData) == 0 {
		return nil, errors.New("no data available")
	}

	// add padding if needed
	// ATMega328 has 64 words per page / 128 bytes per page
	paddingCount := 128 - (len(actualData) % 128)
	for paddingCount > 0 {
		actualData = append(actualData, 255) // 255 => 0xFF
		paddingCount--
	}

	numberOfBlocks := uint16(len(actualData) / blockSize)

	// calculate crc
	// Source: https://github.com/mysensors/MySensorsBootloaderRF24/blob/37dcc50bf2825a2639fe904be8f3309df7b5859e/HW.h#L235
	crc := uint16(0xFFFF)
	for _, b := range actualData {
		crc ^= uint16(b)
		for bit := 0; bit < 8; bit++ {
			crc = (crc >> 1) ^ (-(crc & 1) & 0xA001)
		}
	}

	fw := &firmwareRaw{
		Type:       typeID,
		Version:    versionID,
		Data:       actualData,
		Blocks:     numberOfBlocks,
		CRC:        crc,
		LastAccess: time.Now(),
	}
	return fw, nil
}

func (p *Provider) setNodeLabels(node *nodeTY.Node) {
	busUtils.PostToResourceService(p.logger, p.bus, node.ID, node, rsTY.TypeNode, rsTY.CommandSetLabel, "")
}

// toHex returns hex string
func (p *Provider) toHex(in interface{}) (string, error) {
	var bBuf bytes.Buffer
	err := binary.Write(&bBuf, binary.LittleEndian, in)
	if err != nil {
		return "", err
	}
	return hexENC.EncodeToString(bBuf.Bytes()), nil
}

// toStruct updates struct from hex string
func toStruct(hex string, out interface{}) error {
	hb, err := hexENC.DecodeString(hex)
	if err != nil {
		return err
	}
	r := bytes.NewReader(hb)
	return binary.Read(r, binary.LittleEndian, out)
}

func (p *Provider) updateFirmwareProgressStatus(node *nodeTY.Node, currentBlock, totalBytes int) {
	otaBlockOrder := node.Labels.Get(types.LabelNodeOTABlockOrder)
	if otaBlockOrder == "" {
		otaBlockOrder = OTABlockOrderReverse
	}

	totalBlocks := totalBytes / firmwareBlockSize
	if totalBytes%firmwareBlockSize != 0 {
		totalBlocks++
	}

	lastBlock := totalBlocks - 1

	if currentBlock == 0 ||
		currentBlock%10 == 0 || // number of blocks once send the status
		currentBlock == lastBlock {

		var startTime interface{}
		var endTime interface{}

		var isRunning bool
		percentage := float64(currentBlock) / float64(lastBlock)
		if otaBlockOrder == OTABlockOrderReverse {
			percentage = 1 - percentage
			isRunning = currentBlock != 0
			if currentBlock == lastBlock {
				startTime = time.Now()
			} else if currentBlock == 0 {
				endTime = time.Now()
			}
		} else {
			isRunning = currentBlock != lastBlock
			if currentBlock == 0 {
				startTime = time.Now()
			} else if currentBlock == lastBlock {
				endTime = time.Now()
			}
		}

		// update the status
		state := map[string]interface{}{
			types.FieldOTARunning:     isRunning,
			types.FieldOTAProgress:    int(percentage * 100),
			types.FieldOTAStatusOn:    time.Now(),
			types.FieldOTABlockNumber: currentBlock,
			types.FieldOTAStartTime:   startTime,
			types.FieldOTAEndTime:     endTime,
			types.FieldOTABlockTotal:  totalBlocks,
		}

		// publish the state
		busUtils.PostToResourceService(p.logger, p.bus, node.ID, state, rsTY.TypeNode, rsTY.CommandFirmwareState, "")
	}
}
