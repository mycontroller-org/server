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
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"go.uber.org/zap"
)

// executeFirmwareConfigRequest executes firmware config request and response with hex payload
func (p *Provider) executeFirmwareConfigRequest(msg *msgTY.Message) (string, error) {
	startTime := time.Now()
	rxPL := sanitizeOtaHex(msg.Payloads[0].Value.String())

	// convert the received hex to matching struct format (first 10 bytes when present)
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

	fromReq := otaBlockSizeFromRequest(rxPL)
	if fromReq > 0 {
		p.rememberOtaBlockSize(node, fromReq)
	}

	if isFotaDisabled(node) {
		return p.respondFotaDisabled(node, rxPL, fwCfgReq, startTime)
	}

	if hasFotaScript(node) {
		scriptDisabled, err := p.isFotaScriptDisabled(node)
		if err != nil {
			p.logger.Error("fota script lookup failed", zap.String("nodeId", node.ID), zap.Error(err))
			return "", err
		}
		if scriptDisabled {
			return p.respondFotaScriptInactive(node, rxPL, fwCfgReq, "fota script disabled in data repository")
		}
		scriptRes, err := p.runFotaOnConfig(node, rxPL)
		if err != nil {
			p.logger.Error("fota onConfig failed", zap.String("nodeId", node.ID), zap.Error(err))
			return "", err
		}
		if node.Labels.GetBool(LabelEraseEEPROM) {
			return p.buildFirmwareConfigResponse(node, &firmwareRaw{}, startTime, "")
		}
		if scriptRes.NoUpdate {
			p.logFotaNoUpdate(node, rxPL, fwCfgReq, scriptRes, startTime)
			// UI firmware_update has no CONFIG_REQUEST bytes. Do not invent a
			// type/version/CRC; a mismatch would start an unwanted OTA.
			if rxPL == "" {
				return "", nil
			}
			return p.toHex(&firmwareConfigResponse{
				Type:    fwCfgReq.Type,
				Version: fwCfgReq.Version,
				Blocks:  fwCfgReq.Blocks,
				CRC:     fwCfgReq.CRC,
			})
		}
		if scriptRes.ResponseHex != "" {
			p.logger.Debug("sending firmware config response from fota script",
				zap.String("nodeId", node.ID),
				zap.String("fotaScript", getFotaScriptID(node)),
				zap.String("timeTaken", time.Since(startTime).String()),
			)
			return scriptRes.ResponseHex, nil
		}
		blockSize, err := p.resolveOtaBlockSize(node, rxPL, fromReq, scriptRes.BlockSize)
		if err != nil {
			p.logger.Error("cannot advertise firmware: unknown OTA block size",
				zap.String("nodeId", node.ID),
				zap.String("payload", rxPL),
				zap.Int("parsedBlockSize", fromReq),
				zap.Int("scriptBlockSize", scriptRes.BlockSize),
				zap.Int("labelBlockSize", p.otaBlockSizeFromLabel(node)),
				zap.Error(err),
			)
			return "", err
		}
		fwRaw, err := p.fetchFirmwareByID(scriptRes.FirmwareID, fwCfgReq.Type, fwCfgReq.Version, false, blockSize)
		if err != nil {
			p.logger.Error("error to get firmware from fota onConfig", zap.String("firmwareId", scriptRes.FirmwareID), zap.Error(err))
			return "", err
		}
		fwRaw.LastAccess = time.Now()
		p.rememberFotaSession(node, scriptRes.FirmwareID, fwRaw.BlockSize)
		return p.buildFirmwareConfigResponse(node, fwRaw, startTime, scriptRes.FirmwareID)
	}

	// Stock path: assigned_firmware label (only when fota_script is not set)
	blockSize, err := p.resolveOtaBlockSize(node, rxPL, fromReq, 0)
	if err != nil {
		p.logger.Error("cannot advertise firmware: unknown OTA block size",
			zap.String("nodeId", node.ID),
			zap.String("payload", rxPL),
			zap.Int("parsedBlockSize", fromReq),
			zap.Error(err),
		)
		return "", err
	}
	fwRaw, err := p.fetchFirmware(node, fwCfgReq.Type, fwCfgReq.Version, false, blockSize)
	if err != nil {
		p.logger.Error("error to get firmware", zap.Any("fwCfgReq", fwCfgReq), zap.Error(err))
		return "", err
	}
	fwRaw.LastAccess = time.Now()
	return p.buildFirmwareConfigResponse(node, fwRaw, startTime, node.Labels.Get(types.LabelNodeAssignedFirmware))
}

func (p *Provider) logFotaNoUpdate(node *nodeTY.Node, rxPL string, req *firmwareConfigRequest, scriptRes *fotaScriptResult, startTime time.Time) {
	if p == nil || p.logger == nil {
		return
	}
	fields := []zap.Field{
		zap.String("nodeId", node.ID),
		zap.String("gatewayId", node.GatewayID),
		zap.String("msNodeId", node.NodeID),
		zap.String("fotaScript", getFotaScriptID(node)),
		zap.String("timeTaken", time.Since(startTime).String()),
		zap.String("payload", rxPL),
		zap.Any("request", fotaRequestForLog(rxPL)),
	}
	if req != nil {
		fields = append(fields,
			zap.Uint16("type", req.Type),
			zap.Uint16("version", req.Version),
			zap.Uint16("blocks", req.Blocks),
			zap.Uint16("crc", req.CRC),
			zap.String("crcHex", fmt.Sprintf("%04X", req.CRC)),
		)
	}
	if scriptRes != nil {
		if scriptRes.FirmwareID != "" {
			fields = append(fields, zap.String("firmwareId", scriptRes.FirmwareID))
		}
		if scriptRes.BlockSize > 0 {
			fields = append(fields, zap.Int("blockSize", scriptRes.BlockSize))
		}
		if len(scriptRes.Labels) > 0 {
			fields = append(fields, zap.Any("labels", scriptRes.Labels))
		}
	}
	p.logger.Debug("fota onConfig: no update", fields...)
}

func (p *Provider) respondFotaDisabled(node *nodeTY.Node, rxPL string, req *firmwareConfigRequest, startTime time.Time) (string, error) {
	p.logger.Debug("FOTA disabled for node",
		zap.String("nodeId", node.ID),
		zap.String("msNodeId", node.NodeID),
		zap.String("timeTaken", time.Since(startTime).String()),
	)
	return p.echoFirmwareConfigOrSkip(rxPL, req)
}

func (p *Provider) respondFotaScriptInactive(node *nodeTY.Node, rxPL string, req *firmwareConfigRequest, reason string) (string, error) {
	p.logger.Debug(reason,
		zap.String("nodeId", node.ID),
		zap.String("fotaScript", getFotaScriptID(node)),
	)
	return p.echoFirmwareConfigOrSkip(rxPL, req)
}

func (p *Provider) echoFirmwareConfigOrSkip(rxPL string, req *firmwareConfigRequest) (string, error) {
	if rxPL == "" || req == nil {
		return "", nil
	}
	return p.toHex(&firmwareConfigResponse{
		Type:    req.Type,
		Version: req.Version,
		Blocks:  req.Blocks,
		CRC:     req.CRC,
	})
}

// fotaRequestForLog is parseFotaRequest without raw []byte (JSON would be base64).
func fotaRequestForLog(rxPL string) map[string]interface{} {
	parsed := parseFotaRequest(rxPL, true)
	out := make(map[string]interface{}, len(parsed))
	for k, v := range parsed {
		if _, isBytes := v.([]byte); isBytes {
			continue
		}
		out[k] = v
	}
	return out
}

func (p *Provider) buildFirmwareConfigResponse(node *nodeTY.Node, fwRaw *firmwareRaw, startTime time.Time, fwID string) (string, error) {
	fwCfgRes := &firmwareConfigResponse{}

	// if erase eeprom set for this node, update erase eeprom command and clear the label on the node detail
	if node.Labels.GetBool(LabelEraseEEPROM) {
		p.logger.Debug("erase EEPROM enabled, sending erase EEPROM command to the node", zap.String("nodeId", node.ID))
		fwCfgRes.SetEraseEEPROM()
		node.Labels.Set(LabelEraseEEPROM, "false")
		p.setNodeLabels(node)
	} else {
		fwCfgRes.Type = fwRaw.Type
		fwCfgRes.Version = fwRaw.Version
		fwCfgRes.Blocks = fwRaw.Blocks
		fwCfgRes.CRC = fwRaw.CRC
	}
	p.logger.Debug("sending a firmware config response",
		zap.Any("response", fwCfgRes),
		zap.Int("blockSize", fwRaw.BlockSize),
		zap.String("firmwareId", fwID),
		zap.String("fotaScript", getFotaScriptID(node)),
		zap.String("timeTaken", time.Since(startTime).String()),
	)

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

	// Use the slice size advertised in CONFIG_RESPONSE. Do not re-guess per block.
	blockSize := p.getFotaSessionBlockSize(node)
	if blockSize <= 0 {
		blockSize = p.otaBlockSizeForNode(node)
	}
	if blockSize <= 0 {
		blockSize = defaultFirmwareBlockSize
	}
	if isFotaDisabled(node) {
		p.logger.Debug("FOTA disabled for node",
			zap.String("nodeId", node.ID),
			zap.String("msNodeId", node.NodeID),
		)
		return "", nil
	}

	var fwRaw *firmwareRaw
	if hasFotaScript(node) {
		scriptDisabled, err := p.isFotaScriptDisabled(node)
		if err != nil {
			return "", err
		}
		if scriptDisabled {
			p.logger.Debug("fota script disabled in data repository",
				zap.String("nodeId", node.ID),
				zap.String("fotaScript", getFotaScriptID(node)),
			)
			return "", nil
		}
		scriptRes, err := p.runFotaOnBlock(node, rxPL, fwReq.Type, fwReq.Version, fwReq.Block)
		if err != nil {
			return "", fmt.Errorf("fota onBlock failed: %w", err)
		}
		if scriptRes.ResponseHex != "" {
			p.logger.Debug("sending firmware block response from fota script",
				zap.String("nodeId", node.ID),
				zap.Uint16("block", fwReq.Block),
				zap.String("timeTaken", time.Since(startTime).String()),
			)
			return scriptRes.ResponseHex, nil
		}
		fwRaw, err = p.fetchFirmwareByID(scriptRes.FirmwareID, fwReq.Type, fwReq.Version, false, blockSize)
		if err != nil {
			return "", fmt.Errorf("error on getting firmware from fota onBlock: %w", err)
		}
	} else {
		fwRaw, err = p.fetchFirmware(node, fwReq.Type, fwReq.Version, true, blockSize)
		if err != nil {
			return "", fmt.Errorf("error on getting firmware. %s", err.Error())
		}
	}
	fwRaw.LastAccess = time.Now()

	bs := fwRaw.BlockSize
	if bs <= 0 {
		bs = blockSize
	}
	if bs <= 0 {
		bs = defaultFirmwareBlockSize
	}
	startAddr := int(fwReq.Block) * bs
	endAddr := startAddr + bs
	if endAddr > len(fwRaw.Data) {
		p.logger.Error("requested block is not available", zap.Int("startAddr", startAddr), zap.Int("endAddr", endAddr), zap.Int("maxAvailableAddr", len(fwRaw.Data)))
		return "", fmt.Errorf("requested block is not available: %v", endAddr)
	}
	chunk := fwRaw.Data[startAddr:endAddr]
	if len(chunk) != bs {
		return "", fmt.Errorf("firmware block %d slice length %d != blockSize %d", fwReq.Block, len(chunk), bs)
	}
	hexPL, err := p.encodeFirmwareBlockResponse(fwReq.Type, fwReq.Version, fwReq.Block, chunk)
	if err != nil {
		return "", err
	}
	if len(hexPL) != 2*(6+bs) {
		return "", fmt.Errorf("firmware block hex length %d, want %d (blockSize=%d)", len(hexPL), 2*(6+bs), bs)
	}
	if fwReq.Block == 0 || fwReq.Block%50 == 0 {
		p.logger.Debug("sending a firmware response",
			zap.Any("request", fwReq),
			zap.Int("blockSize", bs),
			zap.Int("responseBytes", 6+bs),
			zap.Int("responseHexLen", len(hexPL)),
			zap.String("timeTaken", time.Since(startTime).String()),
		)
	} else {
		p.logger.Debug("sending a firmware response", zap.Any("request", fwReq), zap.Int("blockSize", bs), zap.String("timeTaken", time.Since(startTime).String()))
	}

	p.updateFirmwareProgressStatus(node, int(fwReq.Block), len(fwRaw.Data), bs)

	return hexPL, nil
}

func packFirmwareBlockResponse(typeID, versionID, block uint16, data []byte) (string, error) {
	buf := make([]byte, 6+len(data))
	binary.LittleEndian.PutUint16(buf[0:2], typeID)
	binary.LittleEndian.PutUint16(buf[2:4], versionID)
	binary.LittleEndian.PutUint16(buf[4:6], block)
	copy(buf[6:], data)
	return hexENC.EncodeToString(buf), nil
}

// encodeFirmwareBlockResponse uses the same 22-byte toHex layout as the
// working 328 path when blockSize is 16.
func (p *Provider) encodeFirmwareBlockResponse(typeID, versionID, block uint16, data []byte) (string, error) {
	if len(data) == defaultFirmwareBlockSize {
		fwRes := firmwareResponse{Type: typeID, Version: versionID, Block: block}
		copy(fwRes.Data[:], data)
		return p.toHex(&fwRes)
	}
	return packFirmwareBlockResponse(typeID, versionID, block, data)
}

func sanitizeOtaHex(requestHex string) string {
	requestHex = strings.TrimSpace(requestHex)
	requestHex = strings.TrimPrefix(requestHex, "0x")
	requestHex = strings.TrimPrefix(requestHex, "0X")
	requestHex = strings.ReplaceAll(requestHex, " ", "")
	requestHex = strings.ReplaceAll(requestHex, "\r", "")
	requestHex = strings.ReplaceAll(requestHex, "\n", "")
	if len(requestHex)%2 == 1 {
		requestHex = requestHex[:len(requestHex)-1]
	}
	return requestHex
}

func validOtaBlockSize(n int) int {
	if n < 8 || n > maxFirmwareBlockSize || n%8 != 0 {
		return 0
	}
	return n
}

func otaBlockSizeFromRequest(requestHex string) int {
	return validOtaBlockSize(advertisedOtaBlockSize(requestHex))
}

// advertisedOtaBlockSize returns the protocol 3.1 size the node reported.
func advertisedOtaBlockSize(requestHex string) int {
	requestHex = sanitizeOtaHex(requestHex)
	if requestHex == "" {
		return 0
	}
	req := parseFotaRequest(requestHex, true)
	if raw, ok := req["blockSize"]; ok {
		n := int(converterUtils.ToInteger(raw))
		if n >= 8 && n%8 == 0 {
			return n
		}
	}
	b, err := hexENC.DecodeString(requestHex)
	if err != nil {
		return 0
	}
	if len(b) >= 11 && binary.LittleEndian.Uint16(b[8:10]) == 0x0103 {
		n := int(b[10])
		if n >= 8 && n%8 == 0 {
			return n
		}
	}
	for i := 0; i+2 < len(b); i++ {
		if b[i] == 0x03 && b[i+1] == 0x01 {
			n := int(b[i+2])
			if n >= 8 && n%8 == 0 {
				return n
			}
		}
	}
	return 0
}

func (p *Provider) rememberOtaBlockSize(node *nodeTY.Node, blockSize int) {
	if node == nil || node.Labels == nil {
		return
	}
	blockSize = validOtaBlockSize(blockSize)
	if blockSize <= 0 {
		return
	}
	want := fmt.Sprintf("%d", blockSize)
	if node.Labels.Get(LabelOtaBlockSize) == want {
		return
	}
	node.Labels.Set(LabelOtaBlockSize, want)
	p.setNodeLabels(node)
}

func (p *Provider) otaBlockSizeFromLabel(node *nodeTY.Node) int {
	if node == nil || node.Labels == nil {
		return 0
	}
	return validOtaBlockSize(node.Labels.GetInt(LabelOtaBlockSize))
}

func (p *Provider) otaBlockSizeForNode(node *nodeTY.Node) int {
	if n := p.getFotaSessionBlockSize(node); n > 0 {
		return n
	}
	n := p.otaBlockSizeFromLabel(node)
	if n > 0 {
		return n
	}
	if hasFotaScript(node) {
		return 0
	}
	return defaultFirmwareBlockSize
}

// resolveOtaBlockSize picks the node's OTA slice size.
// Priority: live CONFIG_REQUEST, script return, trusted label.
// Stock DualOptiboot falls back to 16. Custom FOTA must not advertise 16
// unless the node actually reported 16.
func (p *Provider) resolveOtaBlockSize(node *nodeTY.Node, requestHex string, fromReq, fromScript int) (int, error) {
	if fromReq <= 0 {
		fromReq = otaBlockSizeFromRequest(requestHex)
	}
	fromScript = validOtaBlockSize(fromScript)
	fromLabel := p.otaBlockSizeForNode(node)

	blockSize := fromReq
	if blockSize == 0 {
		blockSize = fromScript
	}
	if blockSize == 0 {
		blockSize = fromLabel
	}
	if blockSize == 0 && !hasFotaScript(node) {
		blockSize = defaultFirmwareBlockSize
	}

	if p.logger != nil {
		p.logger.Debug("resolved OTA block size",
			zap.String("payload", requestHex),
			zap.Int("payloadBytes", len(requestHex)/2),
			zap.Int("fromRequest", fromReq),
			zap.Int("fromScript", fromScript),
			zap.Int("fromLabel", p.otaBlockSizeFromLabel(node)),
			zap.Int("blockSize", blockSize),
		)
	}

	if blockSize == 0 {
		if advertised := advertisedOtaBlockSize(requestHex); advertised > maxFirmwareBlockSize {
			return 0, fmt.Errorf("node advertised OTA blockSize=%d; controller max is %d (raise maxFirmwareBlockSize if radio/MQTT allow it)", advertised, maxFirmwareBlockSize)
		}
		return 0, fmt.Errorf("unknown OTA block size: node did not advertise protocol 3.1 blockSize yet (do not default to 16)")
	}
	if fromReq == 0 && blockSize != defaultFirmwareBlockSize {
		p.rememberOtaBlockSize(node, blockSize)
	}
	return blockSize, nil
}

func firmwareRawCacheKey(fwID string, blockSize int) string {
	return fmt.Sprintf("%s#%d", fwID, blockSize)
}

// fetchFirmware looks requested firmware on memory store (stock path: assigned_firmware label),
// if not available, loads it from disk
func (p *Provider) fetchFirmware(node *nodeTY.Node, typeID, versionID uint16, verifyID bool, blockSize int) (*firmwareRaw, error) {
	fwID := node.Labels.Get(types.LabelNodeAssignedFirmware)
	if fwID == "" {
		return nil, fmt.Errorf("firmware not assigned for this node. gatewayId:%s, nodeId:%s, typeId:%d, versionId:%d",
			node.GatewayID, node.NodeID, typeID, versionID)
	}
	return p.fetchFirmwareByID(fwID, typeID, versionID, verifyID, blockSize)
}

// fetchFirmwareByID loads/caches firmware raw by entity id.
func (p *Provider) fetchFirmwareByID(fwID string, typeID, versionID uint16, verifyID bool, blockSize int) (*firmwareRaw, error) {
	if fwID == "" {
		return nil, fmt.Errorf("firmware id is empty")
	}
	if blockSize <= 0 {
		blockSize = defaultFirmwareBlockSize
	}
	cacheKey := firmwareRawCacheKey(fwID, blockSize)

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

		fwRaw, err := p.getFirmwareRaw(fw.ID, fwTypeID, fwVersionID, blockSize)
		if err != nil {
			p.logger.Error("error on getting firmware data", zap.String("firmwareId", fw.ID), zap.Error(err))
			return nil, err
		}

		// keep it on memory store
		fwRawStore.Add(cacheKey, fwRaw)
		return fwRaw, nil
	}

	// check firmware on memory store
	// if not found, load it from disk
	var fwRaw *firmwareRaw
	fwRawInf := fwRawStore.Get(cacheKey)
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

// isIntelHexFile reports whether file bytes look like Intel HEX (leading ':').
func isIntelHexFile(data []byte) bool {
	for _, b := range data {
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			continue
		}
		return b == ':'
	}
	return false
}

// Source: https://en.wikipedia.org/wiki/Intel_HEX
// https://github.com/mycontroller-org/mycontroller-v1-legacy/blob/1.5.0.Final/modules/core/src/main/java/org/mycontroller/standalone/firmware/FirmwareUtils.java#L118
// https://github.com/mysensors/MySensorsSampleController/blob/9dbae76081a9c080d5fdd68fba9870626025343f/NodeJsController.js#L172
// I8HEX files use only record types 00 and 01 (16-bit addresses)
// 00 - data, 01 - End
//
// Also accepts raw binary (e.g. STM32 signed.bin) when content does not start with ':'.
func (p *Provider) hexByteToLocalFormat(typeID, versionID uint16, hexByte []byte, blockSize int) (*firmwareRaw, error) {
	if len(hexByte) == 0 {
		return nil, errors.New("no data available")
	}
	if blockSize <= 0 {
		return nil, fmt.Errorf("invalid blockSize: %d", blockSize)
	}

	// Raw binary path, not Intel HEX
	if !isIntelHexFile(hexByte) {
		return p.bytesToFirmwareRaw(typeID, versionID, hexByte, blockSize, false)
	}

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

	// DualOptiboot / AVR: pad to 128-byte pages
	return p.bytesToFirmwareRaw(typeID, versionID, actualData, blockSize, true)
}

// bytesToFirmwareRaw builds firmwareRaw from a contiguous image.
// padToAVRPage: when true, pad to 128-byte pages (classic DualOptiboot); when false
// (raw binary), pad only to OTA blockSize with 0xFF.
func (p *Provider) bytesToFirmwareRaw(typeID, versionID uint16, data []byte, blockSize int, padToAVRPage bool) (*firmwareRaw, error) {
	actualData := make([]byte, len(data))
	copy(actualData, data)

	if padToAVRPage {
		// ATMega328 has 64 words per page / 128 bytes per page
		paddingCount := 128 - (len(actualData) % 128)
		for paddingCount > 0 {
			actualData = append(actualData, 255) // 255 => 0xFF
			paddingCount--
		}
	} else if rem := len(actualData) % blockSize; rem != 0 {
		for i := 0; i < blockSize-rem; i++ {
			actualData = append(actualData, 0xFF)
		}
	}

	nBlocks := len(actualData) / blockSize
	if nBlocks > 0xFFFF {
		return nil, fmt.Errorf("image too large for OTA block count: %d bytes (%d blocks)", len(actualData), nBlocks)
	}
	numberOfBlocks := uint16(nBlocks)

	// calculate crc
	// Source: https://github.com/mysensors/MySensorsBootloaderRF24/blob/37dcc50bf2825a2639fe904be8f3309df7b5859e/HW.h#L235
	crc := uint16(0xFFFF)
	for _, b := range actualData {
		crc ^= uint16(b)
		for bit := 0; bit < 8; bit++ {
			crc = (crc >> 1) ^ (-(crc & 1) & 0xA001)
		}
	}

	return &firmwareRaw{
		Type:       typeID,
		Version:    versionID,
		Data:       actualData,
		Blocks:     numberOfBlocks,
		CRC:        crc,
		BlockSize:  blockSize,
		LastAccess: time.Now(),
	}, nil
}

func (p *Provider) setNodeLabels(node *nodeTY.Node) {
	if p == nil || p.bus == nil || node == nil {
		return
	}
	// CommandSetLabel expects label map payload
	busUtils.PostToResourceService(p.logger, p.bus, node.ID, node.Labels, rsTY.TypeNode, rsTY.CommandSetLabel, "")
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

func (p *Provider) updateFirmwareProgressStatus(node *nodeTY.Node, currentBlock, totalBytes, blockSize int) {
	otaBlockOrder := node.Labels.Get(types.LabelNodeOTABlockOrder)
	if otaBlockOrder == "" {
		otaBlockOrder = OTABlockOrderReverse
	}
	if blockSize <= 0 {
		blockSize = defaultFirmwareBlockSize
	}

	totalBlocks := totalBytes / blockSize
	if totalBytes%blockSize != 0 {
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
			switch currentBlock {
			case lastBlock:
				startTime = time.Now()
			case 0:
				endTime = time.Now()
			}
		} else {
			isRunning = currentBlock != lastBlock
			switch currentBlock {
			case 0:
				startTime = time.Now()
			case lastBlock:
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
