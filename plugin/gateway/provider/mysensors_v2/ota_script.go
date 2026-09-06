package mysensors

import (
	"encoding/binary"
	hexENC "encoding/hex"
	"fmt"
	"strings"
	"time"

	repositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/utils/bus_utils/query"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

// Data repository keys under Config.Data for FOTA policy scripts.
const (
	fotaDataKeyOnConfig = "onConfig"
	fotaDataKeyOnBlock  = "onBlock"
	fotaDataKeyDisabled = "disabled"

	// script result map keys
	fotaResultFirmwareID  = "firmwareId"
	fotaResultLabels      = "labels"
	fotaResultResponseHex = "responseHex"
	fotaResultError       = "error"
	fotaResultNoUpdate    = "noUpdate"
	fotaResultBlockSize   = "blockSize"

	// script input variables
	fotaVarRequestHex  = "requestHex"
	fotaVarRequest     = "request"
	fotaVarGatewayID   = "gatewayId"
	fotaVarNodeID      = "nodeId"
	fotaVarNodeLabels  = "nodeLabels"
	fotaVarType        = "type"
	fotaVarVersion     = "version"
	fotaVarBlock       = "block"
	fotaVarGetFirmware = "getFirmware"
	fotaVarCachedFwID  = "cachedFirmwareId"

	defaultFotaScriptTimeout = 5 * time.Second
)

// fotaScriptBundle holds scripts loaded from a data_repository entry.
type fotaScriptBundle struct {
	ID       string
	OnConfig string
	OnBlock  string
	Disabled bool
}

// fotaScriptResult is the normalized return value from onConfig / onBlock.
type fotaScriptResult struct {
	FirmwareID  string
	ResponseHex string
	Labels      map[string]string
	NoUpdate    bool
	BlockSize   int
}

// getFotaScriptID returns trimmed data_repository id from node label fota_script, or "".
func getFotaScriptID(node *nodeTY.Node) string {
	if node == nil || node.Labels == nil {
		return ""
	}
	return strings.TrimSpace(node.Labels.Get(LabelFotaScript))
}

// hasFotaScript reports whether the node has a fota_script label.
func hasFotaScript(node *nodeTY.Node) bool {
	return getFotaScriptID(node) != ""
}

// isFotaDisabled reports node label fota_disabled=true (all OTA off for this node).
func isFotaDisabled(node *nodeTY.Node) bool {
	return node != nil && node.Labels != nil && node.Labels.GetBool(LabelFotaDisabled)
}

// isFotaScriptDisabled is true when the node has a script and the repository has data.disabled.
func (p *Provider) isFotaScriptDisabled(node *nodeTY.Node) (bool, error) {
	if !hasFotaScript(node) {
		return false, nil
	}
	bundle, err := p.getFotaScriptBundle(getFotaScriptID(node))
	if err != nil {
		return false, err
	}
	return bundle.Disabled, nil
}

func bundleFromRepo(cfg *repositoryTY.Config) *fotaScriptBundle {
	if cfg == nil {
		return nil
	}
	return &fotaScriptBundle{
		ID:       cfg.ID,
		OnConfig: strings.TrimSpace(cfg.Data.GetString(fotaDataKeyOnConfig)),
		OnBlock:  strings.TrimSpace(cfg.Data.GetString(fotaDataKeyOnBlock)),
		Disabled: cfg.Data.GetBool(fotaDataKeyDisabled),
	}
}

func (p *Provider) getFotaScriptBundle(id string) (*fotaScriptBundle, error) {
	if id == "" {
		return nil, fmt.Errorf("fota script id is empty")
	}
	if cached := fotaScriptStore.Get(id); cached != nil {
		if b, ok := cached.(*fotaScriptBundle); ok {
			return b, nil
		}
		fotaScriptStore.Remove(id)
	}

	out := &repositoryTY.Config{}
	var bundle *fotaScriptBundle
	addToStore := func(item interface{}) bool {
		cfg, ok := item.(*repositoryTY.Config)
		if !ok {
			p.logger.Error("error on data conversion for data repository",
				zap.String("receivedType", fmt.Sprintf("%T", item)))
			return false
		}
		bundle = bundleFromRepo(cfg)
		if bundle != nil {
			// keep requested id as cache key even if entity id differs
			if bundle.ID == "" {
				bundle.ID = id
			}
			fotaScriptStore.Add(id, bundle)
		}
		return false
	}

	err := query.QueryResource(p.logger, p.bus, id, rsTY.TypeDataRepository, rsTY.CommandGet,
		nil, addToStore, out, queryTimeout)
	if err != nil {
		return nil, fmt.Errorf("load fota script %q: %w", id, err)
	}
	if bundle == nil {
		return nil, fmt.Errorf("fota script %q not found in data repository", id)
	}
	return bundle, nil
}

// runFotaOnConfig executes data.onConfig for ST_FIRMWARE_CONFIG_REQUEST.
func (p *Provider) runFotaOnConfig(node *nodeTY.Node, requestHex string) (*fotaScriptResult, error) {
	result, err := p.runFotaScript(node, true, requestHex, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	if result.NoUpdate {
		p.clearFotaSession(node)
		return result, nil
	}
	if result.FirmwareID != "" {
		p.rememberFotaFirmware(node, result.FirmwareID)
	}
	return result, nil
}

// runFotaOnBlock executes data.onBlock for ST_FIRMWARE_REQUEST.
// onBlock is optional: when empty, the firmwareId cached from onConfig is used.
func (p *Provider) runFotaOnBlock(node *nodeTY.Node, requestHex string, typeID, versionID, block uint16) (*fotaScriptResult, error) {
	scriptID := getFotaScriptID(node)
	if scriptID == "" {
		return nil, fmt.Errorf("fota_script label is empty")
	}
	bundle, err := p.getFotaScriptBundle(scriptID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(bundle.OnBlock) == "" {
		fwID := p.getFotaSessionFirmwareID(node)
		if fwID == "" {
			return nil, fmt.Errorf("data repository %q has empty onBlock and no firmwareId cached from onConfig", scriptID)
		}
		return &fotaScriptResult{FirmwareID: fwID}, nil
	}

	result, err := p.runFotaScript(node, false, requestHex, typeID, versionID, block)
	if err != nil {
		return nil, err
	}
	if result.FirmwareID != "" {
		p.rememberFotaFirmware(node, result.FirmwareID)
	}
	return result, nil
}

func (p *Provider) runFotaScript(node *nodeTY.Node, isConfig bool, requestHex string, typeID, versionID, block uint16) (*fotaScriptResult, error) {
	scriptID := getFotaScriptID(node)
	if scriptID == "" {
		return nil, fmt.Errorf("fota_script label is empty")
	}

	bundle, err := p.getFotaScriptBundle(scriptID)
	if err != nil {
		return nil, err
	}

	script := bundle.OnBlock
	phaseName := fotaDataKeyOnBlock
	if isConfig {
		script = bundle.OnConfig
		phaseName = fotaDataKeyOnConfig
	}
	if strings.TrimSpace(script) == "" {
		return nil, fmt.Errorf("data repository %q has empty %s script", scriptID, phaseName)
	}

	labelsMap := map[string]string{}
	if node.Labels != nil {
		for k, v := range node.Labels {
			labelsMap[k] = v
		}
	}

	variables := map[string]interface{}{
		fotaVarRequestHex: requestHex,
		fotaVarRequest:    parseFotaRequest(requestHex, isConfig),
		fotaVarGatewayID:  node.GatewayID,
		fotaVarNodeID:     node.NodeID,
		fotaVarNodeLabels: labelsMap,
		fotaVarGetFirmware: func(id string) map[string]interface{} {
			return p.firmwareMetaForScript(id, p.otaBlockSizeForNode(node))
		},
		fotaVarCachedFwID: p.getFotaSessionFirmwareID(node),
	}
	if !isConfig {
		variables[fotaVarType] = typeID
		variables[fotaVarVersion] = versionID
		variables[fotaVarBlock] = block
	}

	// Wrap in IIFE so authors can use top-level `return { ... }` (goja forbids bare return).
	wrapped := "(function() {\n" + script + "\n})()"

	timeout := defaultFotaScriptTimeout
	raw, err := javascript.Execute(p.logger, wrapped, variables, &timeout)
	if err != nil {
		return nil, fmt.Errorf("fota script %s (%s): %w", scriptID, phaseName, err)
	}

	result, err := parseFotaScriptResult(raw)
	if err != nil {
		return nil, fmt.Errorf("fota script %s (%s): %w", scriptID, phaseName, err)
	}

	// Apply label updates returned by the script (any keys; BL policy owns them)
	if len(result.Labels) > 0 {
		if node.Labels == nil {
			node.Labels = make(map[string]string)
		}
		changed := false
		for k, v := range result.Labels {
			if node.Labels.Get(k) != v {
				node.Labels.Set(k, v)
				changed = true
			}
		}
		if changed {
			p.setNodeLabels(node)
		}
	}

	return result, nil
}

func (p *Provider) fotaSessionKey(node *nodeTY.Node) string {
	if node == nil {
		return ""
	}
	return p.getNodeStoreID(node.GatewayID, node.NodeID)
}

func (p *Provider) rememberFotaFirmware(node *nodeTY.Node, fwID string) {
	p.rememberFotaSession(node, fwID, 0)
}

func (p *Provider) rememberFotaSession(node *nodeTY.Node, fwID string, blockSize int) {
	key := p.fotaSessionKey(node)
	fwID = strings.TrimSpace(fwID)
	if key == "" || fwID == "" {
		return
	}
	prev := p.getFotaSession(node)
	if blockSize <= 0 && prev != nil {
		blockSize = prev.BlockSize
	}
	fotaSessionStore.Add(key, &fotaSession{FirmwareID: fwID, BlockSize: blockSize})
}

func (p *Provider) getFotaSession(node *nodeTY.Node) *fotaSession {
	key := p.fotaSessionKey(node)
	if key == "" {
		return nil
	}
	v := fotaSessionStore.Get(key)
	if v == nil {
		return nil
	}
	if s, ok := v.(*fotaSession); ok {
		return s
	}
	// older sessions stored a bare firmwareId string
	if id, ok := v.(string); ok && id != "" {
		return &fotaSession{FirmwareID: id}
	}
	return nil
}

func (p *Provider) getFotaSessionFirmwareID(node *nodeTY.Node) string {
	if s := p.getFotaSession(node); s != nil {
		return s.FirmwareID
	}
	return ""
}

func (p *Provider) getFotaSessionBlockSize(node *nodeTY.Node) int {
	if s := p.getFotaSession(node); s != nil {
		return validOtaBlockSize(s.BlockSize)
	}
	return 0
}

func (p *Provider) clearFotaSession(node *nodeTY.Node) {
	key := p.fotaSessionKey(node)
	if key == "" {
		return
	}
	fotaSessionStore.Remove(key)
}

// firmwareMetaForScript is injected into FOTA JS as getFirmware(id).
func (p *Provider) firmwareMetaForScript(id string, blockSize int) map[string]interface{} {
	id = strings.TrimSpace(id)
	if id == "" {
		return map[string]interface{}{"error": "firmware id is empty"}
	}
	fw, err := p.getFirmware(id)
	if err != nil {
		return map[string]interface{}{"id": id, "error": err.Error()}
	}
	typeID := uint16(fw.Labels.GetInt(LabelFirmwareTypeID))
	versionID := uint16(fw.Labels.GetInt(LabelFirmwareVersionID))
	labels := map[string]string{}
	if fw.Labels != nil {
		for k, v := range fw.Labels {
			labels[k] = v
		}
	}
	out := map[string]interface{}{
		"id":       fw.ID,
		"type":     typeID,
		"version":  versionID,
		"labels":   labels,
		"checksum": fw.File.Checksum,
	}
	raw, err := p.getFirmwareRaw(id, typeID, versionID, blockSize)
	if err != nil {
		out["error"] = err.Error()
		return out
	}
	out["crc"] = raw.CRC
	out["blocks"] = raw.Blocks
	n := len(raw.Data)
	if n > fotaFirmwareHeadBytes {
		n = fotaFirmwareHeadBytes
	}
	if n > 0 {
		out["head"] = hexENC.EncodeToString(raw.Data[:n])
	}
	return out
}

// parseFotaRequest exposes the raw hex plus stock / protocol 3.1 fields when present.
// Encoding of img_* (A/B vs mcuboot vs DualOptiboot) stays in the script.
func parseFotaRequest(requestHex string, isConfig bool) map[string]interface{} {
	requestHex = sanitizeOtaHex(requestHex)
	out := map[string]interface{}{
		"hex":    requestHex,
		"length": 0,
	}
	if requestHex == "" {
		return out
	}
	b, err := hexENC.DecodeString(requestHex)
	if err != nil {
		out["error"] = err.Error()
		return out
	}
	out["bytes"] = b
	out["length"] = len(b)
	if isConfig {
		if len(b) >= 10 {
			out["type"] = binary.LittleEndian.Uint16(b[0:2])
			out["version"] = binary.LittleEndian.Uint16(b[2:4])
			out["blocks"] = binary.LittleEndian.Uint16(b[4:6])
			out["crc"] = binary.LittleEndian.Uint16(b[6:8])
			out["blVersion"] = binary.LittleEndian.Uint16(b[8:10])
		}
		if len(b) >= 12 {
			out["blockSize"] = b[10]
			out["imgCommitted"] = b[11]
		}
		if len(b) >= 14 {
			out["imgRevision"] = binary.LittleEndian.Uint16(b[12:14])
		}
		if len(b) >= 18 {
			out["imgBuildNum"] = binary.LittleEndian.Uint32(b[14:18])
		}
		return out
	}
	if len(b) >= 6 {
		out["type"] = binary.LittleEndian.Uint16(b[0:2])
		out["version"] = binary.LittleEndian.Uint16(b[2:4])
		out["block"] = binary.LittleEndian.Uint16(b[4:6])
	}
	return out
}

func parseFotaScriptResult(raw interface{}) (*fotaScriptResult, error) {
	if raw == nil {
		return nil, fmt.Errorf("script returned nil")
	}
	m, err := javascript.ToMap(raw)
	if err != nil {
		return nil, fmt.Errorf("script must return a map/object: %w", err)
	}

	if errMsg := strings.TrimSpace(converterUtils.ToString(m[fotaResultError])); errMsg != "" {
		return nil, fmt.Errorf("%s", errMsg)
	}

	result := &fotaScriptResult{
		FirmwareID:  strings.TrimSpace(converterUtils.ToString(m[fotaResultFirmwareID])),
		ResponseHex: strings.TrimSpace(converterUtils.ToString(m[fotaResultResponseHex])),
		Labels:      map[string]string{},
		NoUpdate:    converterUtils.ToBool(m[fotaResultNoUpdate]),
		BlockSize:   validOtaBlockSize(int(converterUtils.ToInteger(m[fotaResultBlockSize]))),
	}

	if labelsRaw, ok := m[fotaResultLabels]; ok && labelsRaw != nil {
		switch lv := labelsRaw.(type) {
		case map[string]string:
			for k, v := range lv {
				result.Labels[k] = v
			}
		case map[string]interface{}:
			for k, v := range lv {
				result.Labels[k] = converterUtils.ToString(v)
			}
		default:
			if lm, err := javascript.ToMap(labelsRaw); err == nil {
				for k, v := range lm {
					result.Labels[k] = converterUtils.ToString(v)
				}
			}
		}
	}

	if !result.NoUpdate && result.FirmwareID == "" && result.ResponseHex == "" {
		return nil, fmt.Errorf("script must return firmwareId, responseHex, and/or noUpdate")
	}
	return result, nil
}
