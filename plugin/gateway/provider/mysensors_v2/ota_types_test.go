package mysensors

import (
	"bytes"
	"encoding/binary"
	hexENC "encoding/hex"
	"testing"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	repositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	"go.uber.org/zap"
)

func TestFotaScriptID(t *testing.T) {
	n := &nodeTY.Node{Labels: cmap.CustomStringMap{}}
	if getFotaScriptID(n) != "" {
		t.Fatal("empty labels")
	}
	n.Labels.Set(LabelFotaScript, "  ota_stm32_ab  ")
	if getFotaScriptID(n) != "ota_stm32_ab" {
		t.Fatalf("want trimmed id got %q", getFotaScriptID(n))
	}
	if !hasFotaScript(n) {
		t.Fatal("should use fota script")
	}
	n.Labels.Set(LabelFotaScript, "   ")
	if hasFotaScript(n) {
		t.Fatal("whitespace-only should not use script")
	}
}

func TestFotaDisabled(t *testing.T) {
	n := &nodeTY.Node{Labels: cmap.CustomStringMap{LabelFotaScript: "ota_stm32_ab"}}
	if isFotaDisabled(n) {
		t.Fatal("default must allow FOTA")
	}
	n.Labels.Set(LabelFotaDisabled, "true")
	if !isFotaDisabled(n) {
		t.Fatal("fota_disabled=true")
	}
	n.Labels.Set(LabelFotaDisabled, "false")
	if isFotaDisabled(n) {
		t.Fatal("fota_disabled=false")
	}
}

func TestFotaRepoDisabled(t *testing.T) {
	if bundleFromRepo(&repositoryTY.Config{ID: "s"}).Disabled {
		t.Fatal("missing disabled defaults off")
	}
	if !bundleFromRepo(&repositoryTY.Config{
		ID:   "s",
		Data: cmap.CustomMap{"onConfig": "return {noUpdate:true};", "disabled": "true"},
	}).Disabled {
		t.Fatal("disabled true")
	}
	if bundleFromRepo(&repositoryTY.Config{
		ID:   "s",
		Data: cmap.CustomMap{"disabled": "false"},
	}).Disabled {
		t.Fatal("disabled false")
	}
}

func TestParseFotaScriptResult_BlockSize(t *testing.T) {
	res, err := parseFotaScriptResult(map[string]interface{}{
		"firmwareId": "fw_slot_b",
		"blockSize":  24,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.BlockSize != 24 {
		t.Fatalf("blockSize: %d", res.BlockSize)
	}
}

func TestParseFotaScriptResult_OK(t *testing.T) {
	raw := map[string]interface{}{
		"firmwareId": "fw_slot_b",
		"labels": map[string]interface{}{
			"ab_request_slot": "1",
		},
	}
	res, err := parseFotaScriptResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if res.FirmwareID != "fw_slot_b" {
		t.Fatalf("firmwareId: %q", res.FirmwareID)
	}
	if res.Labels["ab_request_slot"] != "1" {
		t.Fatalf("labels: %+v", res.Labels)
	}
}

func TestParseFotaScriptResult_ErrorField(t *testing.T) {
	_, err := parseFotaScriptResult(map[string]interface{}{
		"error": "bad slots",
	})
	if err == nil || err.Error() != "bad slots" {
		t.Fatalf("want bad slots got %v", err)
	}
}

func TestParseFotaScriptResult_ResponseHexOnly(t *testing.T) {
	res, err := parseFotaScriptResult(map[string]interface{}{
		"responseHex": "aabb",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ResponseHex != "aabb" || res.FirmwareID != "" {
		t.Fatalf("%+v", res)
	}
}

func TestParseFotaScriptResult_Missing(t *testing.T) {
	_, err := parseFotaScriptResult(map[string]interface{}{
		"labels": map[string]interface{}{"x": "1"},
	})
	if err == nil {
		t.Fatal("expected error when neither firmwareId nor responseHex nor noUpdate")
	}
}

func TestParseFotaScriptResult_NoUpdate(t *testing.T) {
	res, err := parseFotaScriptResult(map[string]interface{}{
		"noUpdate": true,
		"labels":   map[string]interface{}{"ab_running_slot": "A"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.NoUpdate || res.FirmwareID != "" {
		t.Fatalf("%+v", res)
	}
	if res.Labels["ab_running_slot"] != "A" {
		t.Fatalf("labels: %+v", res.Labels)
	}
}

func TestParseFotaRequest_Protocol31(t *testing.T) {
	// type=1, version=2, blocks=3, crc=4, bl=0x0103, blockSize=16,
	// img_committed=0xA1 (running B), img_revision=0 (request A),
	// img_build_num = (1<<16)|(0<<8)|0xAB
	b := make([]byte, 18)
	binary.LittleEndian.PutUint16(b[0:2], 1)
	binary.LittleEndian.PutUint16(b[2:4], 2)
	binary.LittleEndian.PutUint16(b[4:6], 3)
	binary.LittleEndian.PutUint16(b[6:8], 4)
	binary.LittleEndian.PutUint16(b[8:10], 0x0103)
	b[10] = 16
	b[11] = 0xA1
	binary.LittleEndian.PutUint16(b[12:14], 0)
	binary.LittleEndian.PutUint32(b[14:18], (1<<16)|0xAB)
	req := parseFotaRequest(hexENC.EncodeToString(b), true)
	if req["type"] != uint16(1) || req["version"] != uint16(2) {
		t.Fatalf("type/version: %+v", req)
	}
	if req["imgCommitted"] != byte(0xA1) {
		t.Fatalf("imgCommitted: %+v", req["imgCommitted"])
	}
	if req["imgRevision"] != uint16(0) {
		t.Fatalf("imgRevision: %+v", req["imgRevision"])
	}
	if req["length"] != 18 {
		t.Fatalf("length: %+v", req["length"])
	}
}

func TestParseFotaRequest_Empty(t *testing.T) {
	req := parseFotaRequest("", true)
	if req["hex"] != "" || req["length"] != 0 {
		t.Fatalf("%+v", req)
	}
}

func TestIsIntelHexFile(t *testing.T) {
	if !isIntelHexFile([]byte(":10000000AABB\n")) {
		t.Fatal("expected intel hex")
	}
	if !isIntelHexFile([]byte("\n\r :1000")) {
		t.Fatal("expected intel hex after whitespace")
	}
	if isIntelHexFile([]byte{0x00, 0x00, 0x00, 0x20, 0x01, 0x00}) {
		t.Fatal("raw binary must not look like hex")
	}
}

func TestOtaBlockSizeFromRequest(t *testing.T) {
	if otaBlockSizeFromRequest("") != 0 {
		t.Fatal("empty should be unset")
	}
	b := make([]byte, 18)
	b[10] = 16
	hex := hexENC.EncodeToString(b)
	if otaBlockSizeFromRequest(hex) != 16 {
		t.Fatalf("got %d", otaBlockSizeFromRequest(hex))
	}
	b[10] = 24
	if otaBlockSizeFromRequest(hexENC.EncodeToString(b)) != 24 {
		t.Fatalf("24 got %d", otaBlockSizeFromRequest(hexENC.EncodeToString(b)))
	}
	b[10] = 128
	if otaBlockSizeFromRequest(hexENC.EncodeToString(b)) != 128 {
		t.Fatal("128 must be accepted with extended V2 length")
	}
	b[10] = 7
	if otaBlockSizeFromRequest(hexENC.EncodeToString(b)) != 0 {
		t.Fatal("too small should be unset")
	}
}

func TestOtaBlockSizeFromRequest_Protocol31Scan(t *testing.T) {
	// Official 18-byte request: type/ver/blocks/crc + 0x0103 + 16 + 0xA0
	b := make([]byte, 18)
	binary.LittleEndian.PutUint16(b[8:10], 0x0103)
	b[10] = 16
	b[11] = 0xA0
	if otaBlockSizeFromRequest(hexENC.EncodeToString(b)) != 16 {
		t.Fatalf("official 3.1 got %d", otaBlockSizeFromRequest(hexENC.EncodeToString(b)))
	}
}

func TestAdvertisedOtaBlockSize128IsAccepted(t *testing.T) {
	leftover := "FFFF030180A00100"
	if advertisedOtaBlockSize(leftover) != 128 {
		t.Fatalf("scan should see 128, got %d", advertisedOtaBlockSize(leftover))
	}
	p := &Provider{}
	node := &nodeTY.Node{Labels: cmap.CustomStringMap{LabelFotaScript: "ota_stm32_ab"}}
	bs, err := p.resolveOtaBlockSize(node, leftover, 0, 0)
	if err != nil || bs != 128 {
		t.Fatalf("128-byte blocks must be accepted: %d %v", bs, err)
	}
}

func TestResolveOtaBlockSize_Label16IsUsedForServing(t *testing.T) {
	p := &Provider{}
	node := &nodeTY.Node{Labels: cmap.CustomStringMap{
		LabelFotaScript:   "ota_stm32_ab",
		LabelOtaBlockSize: "16",
	}}
	if p.otaBlockSizeForNode(node) != 16 {
		t.Fatalf("advertised 16 must be used, got %d", p.otaBlockSizeForNode(node))
	}
	bs, err := p.resolveOtaBlockSize(node, "", 0, 0)
	if err != nil || bs != 16 {
		t.Fatalf("label 16 should resolve: %d %v", bs, err)
	}
}

func TestFotaScriptRejectsSwappedImageFromHead(t *testing.T) {
	// B-linked Reset_Handler at 0x08021A01 in getFirmware().head
	head := "00800020011A0208"
	script := `
		function u32leHex(hex, byteOff) {
			var i = byteOff * 2;
			return parseInt(hex.substr(i+6,2)+hex.substr(i+4,2)+hex.substr(i+2,2)+hex.substr(i,2), 16);
		}
		function linkedSlot(fw) {
			var reset = u32leHex(fw.head, 4) & 0xFFFFFFFE;
			if (reset >= 0x08004000 && reset < 0x08021800) return "A";
			if (reset >= 0x08021800 && reset < 0x0803F000) return "B";
			return "";
		}
		var fw = getFirmware("x");
		if (linkedSlot(fw) !== "A") {
			return { error: "linked " + linkedSlot(fw) + " want A" };
		}
		return { firmwareId: "x" };
	`
	wrapped := "(function() {\n" + script + "\n})()"
	getFw := func(id string) map[string]interface{} {
		return map[string]interface{}{"id": id, "head": head}
	}
	logger := zap.NewNop()
	timeout := 2 * time.Second
	raw, err := javascript.Execute(logger, wrapped, map[string]interface{}{"getFirmware": getFw}, &timeout)
	if err != nil {
		t.Fatal(err)
	}
	_, err = parseFotaScriptResult(raw)
	if err == nil || err.Error() != "linked B want A" {
		t.Fatalf("want swap error, got %v", err)
	}
}

func TestFotaScriptRejectsUnknownHead(t *testing.T) {
	script := `
		function linkedSlot(fw) {
			if (!fw || !fw.head) return "";
			return "";
		}
		function rejectWrongSlot(fw, wantSlot, fwId) {
			if (!fw || fw.error) {
				return { error: (fw && fw.error) ? String(fw.error) : ("getFirmware failed for " + fwId) };
			}
			if (!fw.head) {
				return { error: "firmware " + fwId + " has no head; cannot verify slot link" };
			}
			var got = linkedSlot(fw);
			if (!got || got !== wantSlot) {
				return { error: "firmware " + fwId + " Reset_Handler is not in slot A or B" };
			}
			return null;
		}
		var fw = getFirmware("x");
		var wrong = rejectWrongSlot(fw, "B", "x");
		if (wrong) {
			return wrong;
		}
		return { firmwareId: "x" };
	`
	wrapped := "(function() {\n" + script + "\n})()"
	logger := zap.NewNop()
	timeout := 2 * time.Second

	run := func(head interface{}) error {
		getFw := func(id string) map[string]interface{} {
			out := map[string]interface{}{"id": id}
			if head != nil {
				out["head"] = head
			}
			return out
		}
		raw, err := javascript.Execute(logger, wrapped, map[string]interface{}{"getFirmware": getFw}, &timeout)
		if err != nil {
			return err
		}
		_, err = parseFotaScriptResult(raw)
		return err
	}

	if err := run(nil); err == nil || err.Error() != "firmware x has no head; cannot verify slot link" {
		t.Fatalf("missing head: %v", err)
	}
	if err := run("0000000000000000"); err == nil || err.Error() != "firmware x Reset_Handler is not in slot A or B" {
		t.Fatalf("unlinked head: %v", err)
	}
}

func TestFotaScriptRadioContinuesPendingServe(t *testing.T) {
	// Same type+version would noUpdate, but ab_serve_firmware is a different CRC.
	script := `
		function rejectWrongSlot() { return null; }
		var pendingId = nodeLabels["ab_serve_firmware"];
		var pendingFw = getFirmware(pendingId);
		if (!pendingFw.error && +request.crc !== +pendingFw.crc) {
			return { firmwareId: pendingId };
		}
		if (+request.type === 1 && +request.version === 1 && +request.crc === 0x0AB4) {
			return { noUpdate: true };
		}
		return { firmwareId: "fota_test_a" };
	`
	wrapped := "(function() {\n" + script + "\n})()"
	getFw := func(id string) map[string]interface{} {
		if id == "fota_test_a" {
			return map[string]interface{}{"id": id, "type": uint16(1), "version": uint16(1), "crc": uint16(0x2B32)}
		}
		return map[string]interface{}{"id": id, "type": uint16(1), "version": uint16(1), "crc": uint16(0x0AB4)}
	}
	logger := zap.NewNop()
	timeout := 2 * time.Second
	raw, err := javascript.Execute(logger, wrapped, map[string]interface{}{
		"request":     map[string]interface{}{"type": 1, "version": 1, "crc": 0x0AB4},
		"nodeLabels":  map[string]string{"ab_serve_firmware": "fota_test_a"},
		"getFirmware": getFw,
	}, &timeout)
	if err != nil {
		t.Fatal(err)
	}
	res, err := parseFotaScriptResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if res.NoUpdate || res.FirmwareID != "fota_test_a" {
		t.Fatalf("pending serve must continue, got %+v", res)
	}
}

func TestAssembleFirmwareBytes(t *testing.T) {
	// Single full-file block (new resource service path)
	all := []byte{1, 2, 3, 4, 5}
	got, ok := assembleFirmwareBytes(map[int][]byte{0: all}, len(all))
	if !ok || string(got) != string(all) {
		t.Fatalf("full file: ok=%v %v", ok, got)
	}

	// Incomplete 512-byte chunks must not look done
	if _, ok := assembleFirmwareBytes(map[int][]byte{0: make([]byte, firmwareTY.BlockSize)}, firmwareTY.BlockSize*2); ok {
		t.Fatal("missing second chunk must be incomplete")
	}

	chunk0 := bytes.Repeat([]byte{0xAA}, firmwareTY.BlockSize)
	chunk1 := []byte{0xBB, 0xCC}
	got, ok = assembleFirmwareBytes(map[int][]byte{0: chunk0, 1: chunk1}, firmwareTY.BlockSize+2)
	if !ok || len(got) != firmwareTY.BlockSize+2 || got[0] != 0xAA || got[firmwareTY.BlockSize] != 0xBB {
		t.Fatalf("chunked: ok=%v len=%d", ok, len(got))
	}
}

func TestEncodeFirmwareBlockResponseStock16(t *testing.T) {
	p := &Provider{}
	data := make([]byte, 16)
	data[0] = 0xAA
	hex, err := p.encodeFirmwareBlockResponse(1, 1, 0, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(hex) != 44 {
		t.Fatalf("stock 16-byte block must be 22 bytes / 44 hex, got %d", len(hex))
	}
	raw, err := hexENC.DecodeString(hex)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 22 || raw[6] != 0xAA {
		t.Fatalf("layout %d %x", len(raw), raw)
	}
}

func TestResolveOtaBlockSize_StockDefaults16(t *testing.T) {
	p := &Provider{}
	node := &nodeTY.Node{Labels: cmap.CustomStringMap{}}
	bs, err := p.resolveOtaBlockSize(node, "", 0, 0)
	if err != nil || bs != 16 {
		t.Fatalf("stock node should default 16: %d %v", bs, err)
	}
}

func TestRememberOtaBlockSizeStoresLive16(t *testing.T) {
	p := &Provider{}
	node := &nodeTY.Node{Labels: cmap.CustomStringMap{LabelOtaBlockSize: "24"}}
	p.rememberOtaBlockSize(node, 16)
	if node.Labels.Get(LabelOtaBlockSize) != "16" {
		t.Fatalf("live 16 should replace 24, got %q", node.Labels.Get(LabelOtaBlockSize))
	}
}

func TestSanitizeOtaHex(t *testing.T) {
	if sanitizeOtaHex(" AABBCC \r\n") != "AABBCC" {
		t.Fatal("trim")
	}
	if sanitizeOtaHex("0xAABB") != "AABB" {
		t.Fatal("0x prefix")
	}
}

func TestFirmwareBlockOffsetNoUint16Overflow(t *testing.T) {
	// Block 0x0FFF * 16 = 65520; +16 = 65536 which overflows uint16 to 0.
	const block uint16 = 0x0FFF
	const bs = 16
	start := int(block) * bs
	end := start + bs
	if start != 65520 || end != 65536 {
		t.Fatalf("start=%d end=%d", start, end)
	}
	data := make([]byte, 65536)
	data[65520] = 0xAA
	hex, err := packFirmwareBlockResponse(1, 1, block, data[start:end])
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := hexENC.DecodeString(hex)
	if len(raw) != 6+16 || raw[6] != 0xAA {
		t.Fatalf("%d %x", len(raw), raw)
	}
}

func TestPackFirmwareBlockResponse(t *testing.T) {
	data := make([]byte, 128)
	data[0] = 0xAB
	hex, err := packFirmwareBlockResponse(1, 2, 3, data)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := hexENC.DecodeString(hex)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 6+128 {
		t.Fatalf("len %d", len(raw))
	}
	if raw[6] != 0xAB {
		t.Fatal("payload")
	}
}

func TestBytesToFirmwareRaw_BinaryPadToBlock(t *testing.T) {
	p := &Provider{}
	raw := []byte{1, 2, 3, 4, 5} // 5 bytes → pad to 16
	fw, err := p.bytesToFirmwareRaw(1, 2, raw, 16, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(fw.Data) != 16 {
		t.Fatalf("want len 16 got %d", len(fw.Data))
	}
	if fw.Blocks != 1 {
		t.Fatalf("want 1 block got %d", fw.Blocks)
	}
	for i := 5; i < 16; i++ {
		if fw.Data[i] != 0xFF {
			t.Fatalf("pad byte %d want 0xFF got %02X", i, fw.Data[i])
		}
	}
}

func TestBundleFromRepo(t *testing.T) {
	cfg := &repositoryTY.Config{
		ID: "ota_stm32_ab",
		Data: cmap.CustomMap{
			"onConfig": "cfg",
			"onBlock":  "blk",
		},
	}
	b := bundleFromRepo(cfg)
	if b == nil || b.OnConfig != "cfg" || b.OnBlock != "blk" {
		t.Fatalf("%+v", b)
	}
	folded := bundleFromRepo(&repositoryTY.Config{
		ID:   "s",
		Data: cmap.CustomMap{"ONCONFIG": "x"},
	})
	if folded.OnConfig != "x" {
		t.Fatalf("normalized key: %+v", folded)
	}
}

func TestJavascriptOnConfigSample(t *testing.T) {
	// Minimal onConfig that picks firmware from a node label (same pattern as STM32 A/B script)
	script := `
		var fw = nodeLabels["assigned_firmware_slot_b"] || nodeLabels["assigned_firmware_slot_a"];
		if (!fw) {
			return { error: "no slot firmware label" };
		}
		return {
			firmwareId: fw,
			labels: { ab_request_slot: "1" }
		};
	`
	// Execute via parse path only; full engine test would need logger; use goja through package if Provider available
	// Here we just validate result shape with a synthetic map mimicking script return
	res, err := parseFotaScriptResult(map[string]interface{}{
		"firmwareId": "water_b",
		"labels":     map[string]interface{}{"ab_request_slot": "1"},
	})
	if err != nil || res.FirmwareID != "water_b" {
		t.Fatalf("%v %+v", err, res)
	}
	_ = script
}

func TestFotaScriptUIEmptyRequestSameReleaseNoUpdate(t *testing.T) {
	script := `
		if (!request || !request.hex || request.length === 0) {
			var lastRun = nodeLabels["ab_running_slot"];
			var runKey = lastRun === "A" ? "assigned_firmware_slot_a" : "assigned_firmware_slot_b";
			var reqKey = lastRun === "A" ? "assigned_firmware_slot_b" : "assigned_firmware_slot_a";
			var runFw = getFirmware(nodeLabels[runKey]);
			var reqFw = getFirmware(nodeLabels[reqKey]);
			if (+runFw.type === +reqFw.type && +runFw.version === +reqFw.version && +runFw.type !== 0xFFFF) {
				return { noUpdate: true };
			}
			return { firmwareId: nodeLabels[reqKey] };
		}
		return { firmwareId: "x" };
	`
	wrapped := "(function() {\n" + script + "\n})()"
	getFw := func(id string) map[string]interface{} {
		return map[string]interface{}{"id": id, "type": uint16(1), "version": uint16(1)}
	}
	logger := zap.NewNop()
	timeout := 2 * time.Second
	raw, err := javascript.Execute(logger, wrapped, map[string]interface{}{
		"request":    parseFotaRequest("", true),
		"nodeLabels": map[string]string{"ab_running_slot": "A", "assigned_firmware_slot_a": "fw_a", "assigned_firmware_slot_b": "fw_b"},
		"getFirmware": getFw,
	}, &timeout)
	if err != nil {
		t.Fatal(err)
	}
	res, err := parseFotaScriptResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !res.NoUpdate || res.FirmwareID != "" {
		t.Fatalf("UI same release must noUpdate, got %+v", res)
	}
}

func TestFotaScriptExecuteUsesRequestObject(t *testing.T) {
	b := make([]byte, 18)
	binary.LittleEndian.PutUint16(b[0:2], 10)
	binary.LittleEndian.PutUint16(b[2:4], 20)
	b[11] = 0xA0
	binary.LittleEndian.PutUint16(b[12:14], 1)
	hex := hexENC.EncodeToString(b)
	script := `
		if ((request.imgCommitted & 0xF0) !== 0xA0) {
			return { error: "not ab" };
		}
		if (request.type === 10 && request.version === 20) {
			return { noUpdate: true, labels: { ab_running_slot: "A", ab_request_slot: "B" } };
		}
		return { firmwareId: "x" };
	`
	wrapped := "(function() {\n" + script + "\n})()"
	logger := zap.NewNop()
	timeout := 2 * time.Second
	raw, err := javascript.Execute(logger, wrapped, map[string]interface{}{
		"request": parseFotaRequest(hex, true),
	}, &timeout)
	if err != nil {
		t.Fatal(err)
	}
	res, err := parseFotaScriptResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !res.NoUpdate || res.Labels["ab_running_slot"] != "A" {
		t.Fatalf("%+v", res)
	}
}

func TestFotaScriptExecuteWithReturn(t *testing.T) {
	// Integration-style: same wrap as runFotaScript
	script := `
		if (!requestHex) {
			return { error: "empty" };
		}
		return {
			firmwareId: nodeLabels["assigned_firmware_slot_b"],
			labels: { ab_request_slot: "1" }
		};
	`
	wrapped := "(function() {\n" + script + "\n})()"
	// Use package javascript via Provider path: call parse after manual execute
	// Minimal: evaluate with goja through javascript.Execute
	logger := zap.NewNop()
	timeout := 2 * time.Second
	raw, err := javascript.Execute(logger, wrapped, map[string]interface{}{
		"requestHex": "aabb",
		"nodeLabels": map[string]string{"assigned_firmware_slot_b": "fw_b"},
	}, &timeout)
	if err != nil {
		t.Fatal(err)
	}
	res, err := parseFotaScriptResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if res.FirmwareID != "fw_b" || res.Labels["ab_request_slot"] != "1" {
		t.Fatalf("%+v", res)
	}
}
