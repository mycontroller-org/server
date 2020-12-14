package mysensors

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	fwAPI "github.com/mycontroller-org/backend/v2/pkg/api/firmware"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

// executeFirmwareConfigRequest executes firmware config request and response with hex payload
func executeFirmwareConfigRequest(msg *msgml.Message) (string, error) {
	startTime := time.Now()
	rxPL := msg.Payloads[0].Value

	// convert the received hex to matching struct format
	fwCfgReq := &firmwareConfigRequest{}
	err := toStruct(rxPL, fwCfgReq)
	if err != nil {
		return "", err
	}

	// get the node details
	node, err := nodeAPI.GetByIDs(msg.GatewayID, msg.NodeID)
	if err != nil {
		return "", err
	}

	// get firmware raw format
	fwRaw, err := fetchFirmware(node, fwCfgReq.Type, fwCfgReq.Version, false)
	if err != nil {
		return "", err
	}
	fwRaw.LastAccess = time.Now()

	// create firmware config response struct and update required values
	fwCfgRes := &firmwareConfigResponse{}

	// if erase eeprom set for this node, update erase eeprom command and clear the label on the node detail
	if node.Labels.GetBool(LabelEraseEEPROM) {
		zap.L().Debug("Erase EEPROM enabled, sending erase EEPROM command to the node", zap.String("nodeId", node.ID))
		// set erase command
		fwCfgRes.SetEraseEEPROM()
		// remove erase config data from node
		node.Labels.Set(LabelEraseEEPROM, "false")
		err = nodeAPI.Save(node)
		if err != nil {
			return "", err
		}
	} else { // update assigned firmware config details
		fwCfgRes.Type = fwRaw.Type
		fwCfgRes.Version = fwRaw.Version
		fwCfgRes.Blocks = fwRaw.Blocks
		fwCfgRes.CRC = fwRaw.CRC
	}
	zap.L().Info("Sending a firmware config respose", zap.Any("request", fwCfgReq), zap.Any("response", fwCfgRes), zap.String("timeTaken", time.Since(startTime).String()))

	// convert the struct to hex string and return
	return toHex(fwCfgRes)
}

// executeFirmwareRequest executes firmware request and response with hex payload
func executeFirmwareRequest(msg *msgml.Message) (string, error) {
	rxPL := msg.Payloads[0].Value
	startTime := time.Now()

	// convert the received hex to matching struct format
	fwReq := &firmwareRequest{}
	err := toStruct(rxPL, fwReq)
	if err != nil {
		return "", err
	}

	// get the node details
	node, err := nodeAPI.GetByIDs(msg.GatewayID, msg.NodeID)
	if err != nil {
		return "", err
	}

	// get firmware raw format
	fwRaw, err := fetchFirmware(node, fwReq.Type, fwReq.Version, true)
	if err != nil {
		return "", err
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
	copy(fwRes.Data[:], fwRaw.Data[startAddr:(endAddr+1)])
	zap.L().Info("Sending a firmware respose", zap.Any("request", fwReq), zap.Any("response", fwRes), zap.String("timeTaken", time.Since(startTime).String()))

	// convert the struct to hex string and return
	return toHex(fwRes)
}

// fetchFirmware looks requested firmware on memory store,
// if not available, loads it from disk
func fetchFirmware(node *nml.Node, typeID, versionID uint16, verifyID bool) (*firmwareRaw, error) {
	// get mapped firmware by id
	fwID := node.Labels.Get(ml.LabelNodeAssignedFirmware)
	if fwID == "" {
		return nil, errors.New("Firmware not assigned for this node")
	}

	// lambda function to load firmware
	loadFirmwareRawFn := func() (*firmwareRaw, error) {
		// get firmware details
		fw, err := fwAPI.GetByID(fwID)
		if err != nil {
			return nil, err
		}

		// get mysensor specific ids
		if fw.Labels.Get(LabelFirmwareTypeID) == "" || fw.Labels.Get(LabelFirmwareVersionID) == "" {
			return nil, errors.New("Firmware type id or version id not set")
		}
		fwTypeID := uint16(fw.Labels.GetInt(LabelFirmwareTypeID))
		fwVersionID := uint16(fw.Labels.GetInt(LabelFirmwareVersionID))

		// get firmware hex file
		hexFile, err := ut.ReadFile(ml.GetDirectoryFirmware(), fw.File.Name)
		if err != nil {
			return nil, err
		}

		// convert the hex file to raw format
		fwRaw, err := hexByteToLocalFormat(fwTypeID, fwVersionID, hexFile, firmwareBlockSize)
		if err != nil {
			return nil, err
		}

		// keep it on memory store
		fwStore.add(fwID, fwRaw)
		return fwRaw, nil
	}

	// check firmware on memory store
	// if not found, load it from disk
	fwRaw, found := fwStore.get(fwID)
	if !found {
		_fwRaw, err := loadFirmwareRawFn()
		if err != nil {
			return nil, err
		}
		fwRaw = _fwRaw
	}
	if verifyID { // verify firmware ids
		if fwRaw.Type != typeID || fwRaw.Version != versionID {
			return nil, fmt.Errorf("Requested firmware type id or version id not matching[Req, Avl], TypeId:[%v, %v], VersionId:[%v, %v]",
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
//  :10010000214601360121470136007EFE09D2190140
//  :100110002146017E17C20001FF5F16002148011928
//  :10012000194E79234623965778239EDA3F01B2CAA7
//  :100130003F0156702B5E712B722B732146013421C7
//  :00000001FF
//  :(start) xx(byte count) xxxx(address) xx(record type) xxx...xx(data, checksum)
func hexByteToLocalFormat(typeID, versionID uint16, hexByte []byte, blockSize int) (*firmwareRaw, error) {
	hexString := string(hexByte)
	hexString = strings.ReplaceAll(hexString, "\r", "") // remove all "\r" char
	hexLines := strings.Split(hexString, "\n")          // split as separate lines

	actualData := make([]byte, 0)
	for _, line := range hexLines {
		line = strings.TrimSpace(line) // remove spaces if any
		if len(line) == 0 {
			continue
		}
		// first char of the line should be ':'
		if line[0] != ':' {
			return nil, errors.New("hex line not started with the char ':'")
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
		dataBytes, err := hex.DecodeString(data)
		if err != nil {
			zap.L().Error("failed", zap.Any("data", data), zap.Error(err))
			return nil, err
		}
		// include it to our main slice
		actualData = append(actualData, dataBytes...)
	}
	// check the processed bytes length
	if len(actualData) == 0 {
		return nil, errors.New("No data available")
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
