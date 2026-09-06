# Sample FOTA script: STM32 A/B (data repository)

MyController stays generic. This repository entry is the A/B *policy*.
A different bootloader only needs a different repository + node label.

## Setup

1. Create a **Data repository** with id e.g. `ota_stm32_ab`.
2. Put the `onConfig` script below in `data.onConfig`.
   `data.onBlock` is **optional**; omit it. Core reuses the `firmwareId` chosen in `onConfig`.
3. On the node, set labels:
   - `fota_script` = `ota_stm32_ab` (this repository id)
   - `assigned_firmware_slot_a` = firmware entity id of the **slot A** signed image
   - `assigned_firmware_slot_b` = firmware entity id of the **slot B** signed image
4. On each firmware entity, set label `ms_flash_slot` = `A` or `B` (must match how that `.signed.bin` was linked). The script refuses a swap, and also refuses if `getFirmware().head` is missing or the Reset_Handler is not in slot A or B.
5. Leave `assigned_firmware` unused. While `fota_script` is set, stock
   `assigned_firmware` is never used, even if the repository is missing or
   `data.disabled` is true.
6. To turn the script off for every node that points at this repository, set
   `data.disabled` = `true`. CONFIG_REQUEST is echoed (no update). A missing
   repository is an error (still no stock fallback).
7. To turn **all** OTA off for one node (script and stock), set node label
   `fota_disabled` = `true`. CONFIG_REQUEST is echoed (no update); block
   requests and UI firmware update fail.

Both slot images of **one release** must share the same MySensors `ms_type_id` / `ms_version_id`
(different binaries / CRC, same type+version). That is what stops A↔B ping-pong:
if the node already reports that type+version, the script returns `noUpdate`.

When you ship a new release, update **both** slot labels together.

## Protocol (device side)

`ST_FIRMWARE_CONFIG_REQUEST` protocol 3.1 payload (18 bytes), filled by `InternalOtaFlash`:

| Offset | Field | A/B meaning |
|--------|--------|-------------|
| 0–1 | `type` | Running (or staged, if pending confirm) firmware type |
| 2–3 | `version` | Running (or staged) firmware version |
| 11 | `img_committed` | `0xA0` = running **A**, `0xA1` = running **B** |
| 12–13 LE | `img_revision` | Request slot: `0` = serve A, `1` = serve B (inactive) |
| 14–17 | `img_build_num` | `(running<<16) \| (request<<8) \| 0xAB` |

Always OTA the **request (inactive)** slot image, never the running slot.

## Script inputs

| Variable | onConfig | onBlock | Description |
|----------|----------|---------|-------------|
| `request` | yes | yes | Parsed payload: `hex`, `bytes`, `length`, and when present `type`, `version`, `blocks`, `crc`, `blVersion`, `blockSize`, `imgCommitted`, `imgRevision`, `imgBuildNum` (config) or `block` (block request) |
| `requestHex` | yes | yes | Same raw hex as `request.hex` |
| `gatewayId` | yes | yes | Gateway id |
| `nodeId` | yes | yes | MySensors node id |
| `nodeLabels` | yes | yes | Current node labels (keys are lowercase) |
| `getFirmware(id)` | yes | yes | `{id, type, version, labels, checksum, crc, blocks, head}`. `head` is the first 16 image bytes as hex; this script parses the vector table from it |
| `cachedFirmwareId` | yes | yes | `firmwareId` remembered from the last `onConfig` |
| `type` / `version` / `block` | no | yes | From the block request |

Helpers: `mcUtils.convert.HexStringToBytes`, `ToUInt16LE`, etc. Prefer `request.*`.

## Script return (object)

| Field | Required | Description |
|-------|----------|-------------|
| `firmwareId` | one of* | Firmware entity id to serve |
| `responseHex` | one of* | Full hex response; skips core packing |
| `noUpdate` | one of* | Echo the node's current type/version/blocks/crc (no OTA) |
| `blockSize` | no | Node OTA slice size in bytes (8–192, multiple of 8). Core prefers the live CONFIG_REQUEST; use this for UI-triggered empty requests |
| `labels` | no | Merged onto the node. This script sets `ab_running_slot`, `ab_request_slot`, `ab_last_*` (what the node reported, hex16) and `ab_serve_*` (image being offered) |
| `error` | no | Non-empty string aborts OTA |

\* At least one of `firmwareId`, `responseHex`, or `noUpdate`.

`onBlock` may be omitted. If present it runs **on every block**; keep it cheap or leave it empty.

## data.onConfig

```javascript
function slotLetter(n) {
  return n === 0 ? "A" : "B";
}

function slotLabels(runSlot, reqSlot) {
  return {
    ab_capable: "true",
    ab_running_slot: slotLetter(runSlot),
    ab_request_slot: slotLetter(reqSlot)
  };
}

function hex16(n) {
  var v = (+n) & 0xFFFF;
  var s = v.toString(16).toUpperCase();
  while (s.length < 4) {
    s = "0" + s;
  }
  return s;
}

function noteNodeRequest(labels, req) {
  if (!req) {
    return labels;
  }
  if (req.type !== undefined) {
    labels.ab_last_type = hex16(req.type);
  }
  if (req.version !== undefined) {
    labels.ab_last_version = hex16(req.version);
  }
  if (req.crc !== undefined) {
    labels.ab_last_crc = hex16(req.crc);
  }
  return labels;
}

function noteServe(labels, fw, fwId) {
  labels.ab_serve_firmware = fwId || "";
  if (fw && !fw.error) {
    labels.ab_serve_type = hex16(fw.type);
    labels.ab_serve_version = hex16(fw.version);
    if (fw.crc !== undefined) {
      labels.ab_serve_crc = hex16(fw.crc);
    }
  }
  return labels;
}

function pickSlotFirmware(reqSlot) {
  var fwKey = reqSlot === 0 ? "assigned_firmware_slot_a" : "assigned_firmware_slot_b";
  var fwId = nodeLabels[fwKey];
  if (!fwId) {
    return { error: "missing node label " + fwKey };
  }
  return { firmwareId: fwId, fwKey: fwKey };
}

// Little-endian u32 from getFirmware().head (hex). Custom to this bootloader.
function u32leHex(hex, byteOff) {
  var i = byteOff * 2;
  if (!hex || hex.length < i + 8) {
    return 0;
  }
  return parseInt(hex.substr(i + 6, 2) + hex.substr(i + 4, 2) + hex.substr(i + 2, 2) + hex.substr(i, 2), 16);
}

// Slot this image is *linked* for, from the Cortex-M vector table (Reset_Handler).
// Do not trust ms_flash_slot; that is only what was typed in the UI.
function linkedSlot(fw) {
  var SLOT_A = 0x08004000;
  var SLOT_B = 0x08021800;
  var SLOT_SIZE = 0x1D800;
  var reset = u32leHex(fw.head, 4) & 0xFFFFFFFE;
  if (reset >= SLOT_A && reset < SLOT_A + SLOT_SIZE) {
    return "A";
  }
  if (reset >= SLOT_B && reset < SLOT_B + SLOT_SIZE) {
    return "B";
  }
  return "";
}

function rejectWrongSlot(fw, wantSlot, fwId, labels) {
  if (!fw || fw.error) {
    return { error: (fw && fw.error) ? String(fw.error) : ("getFirmware failed for " + fwId), labels: labels };
  }
  if (!fw.head) {
    return { error: "firmware " + fwId + " has no head; cannot verify slot link", labels: labels };
  }
  var got = linkedSlot(fw);
  if (!got) {
    return { error: "firmware " + fwId + " Reset_Handler is not in slot A or B", labels: labels };
  }
  if (got !== wantSlot) {
    return { error: "firmware " + fwId + " is linked for slot " + got + ", node requested " + wantSlot, labels: labels };
  }
  return null;
}

function advertisedBlockSize() {
  var fromReq = request && request.blockSize;
  if (fromReq >= 8 && fromReq <= 192 && (fromReq % 8) === 0) {
    return fromReq;
  }
  var fromLabel = parseInt(nodeLabels["ms_ota_block_size"], 10);
  if (fromLabel >= 8 && fromLabel <= 192 && (fromLabel % 8) === 0 && fromLabel !== 16) {
    return fromLabel;
  }
  return 0;
}

// UI-triggered config request has no payload: use last known running slot if we have it.
if (!request || !request.hex || request.length === 0) {
  var lastRun = nodeLabels["ab_running_slot"];
  if (lastRun !== "A" && lastRun !== "B") {
    return { error: "empty config request and unknown ab_running_slot; wait for the node to present" };
  }
  var uiBs = advertisedBlockSize();
  if (!uiBs) {
    return { error: "empty config request and unknown OTA block size; wait for the node to present protocol 3.1" };
  }
  var uiReqSlot = lastRun === "A" ? 1 : 0;
  var uiPick = pickSlotFirmware(uiReqSlot);
  if (uiPick.error) {
    return { error: uiPick.error };
  }
  var uiReqFw = getFirmware(uiPick.firmwareId);
  var uiLabels = slotLabels(lastRun === "A" ? 0 : 1, uiReqSlot);
  if (uiReqFw.error) {
    return { error: uiReqFw.error, labels: uiLabels };
  }
  var uiWrong = rejectWrongSlot(uiReqFw, slotLetter(uiReqSlot), uiPick.firmwareId, uiLabels);
  if (uiWrong) {
    return uiWrong;
  }
  // UI Firmware update is explicit: offer the inactive slot. Radio
  // CONFIG_REQUEST still skips when the node CRC matches (below).
  return { firmwareId: uiPick.firmwareId, blockSize: uiBs, labels: noteServe(uiLabels, uiReqFw, uiPick.firmwareId) };
}

if (request.length < 18) {
  return { error: "need protocol 3.1 (18 bytes) for A/B, got " + request.length };
}

var imgCommitted = request.imgCommitted;
var reqSlot = request.imgRevision & 0xFF;
if ((imgCommitted & 0xF0) !== 0xA0) {
  return { error: "not an A/B node (imgCommitted high nibble != 0xA0)" };
}
var runSlot = imgCommitted & 0x0F;
if (runSlot > 1 || reqSlot > 1 || runSlot === reqSlot) {
  return { error: "invalid slots run=" + runSlot + " req=" + reqSlot };
}

var pick = pickSlotFirmware(reqSlot);
if (pick.error) {
  return { error: pick.error };
}

var labels = noteNodeRequest(slotLabels(runSlot, reqSlot), request);

var fw = getFirmware(pick.firmwareId);
if (fw.error) {
  return { error: fw.error, labels: labels };
}

var wantSlot = slotLetter(reqSlot);
var wrong = rejectWrongSlot(fw, wantSlot, pick.firmwareId, labels);
if (wrong) {
  return wrong;
}

var runKey = runSlot === 0 ? "assigned_firmware_slot_a" : "assigned_firmware_slot_b";
var runFw = getFirmware(nodeLabels[runKey]);
var bs = advertisedBlockSize();

// A UI click stores ab_serve_firmware. Keep offering that file on reboot
// CONFIG_REQUEST until the node reports that CRC (OTA finished / confirmed).
var pendingId = nodeLabels["ab_serve_firmware"];
if (pendingId) {
  var pendingFw = getFirmware(pendingId);
  if (!pendingFw.error && pendingFw.crc !== undefined && +request.crc !== +pendingFw.crc) {
    var pendingWrong = rejectWrongSlot(pendingFw, wantSlot, pendingId, labels);
    if (!pendingWrong) {
      return { firmwareId: pendingId, blockSize: bs, labels: noteServe(labels, pendingFw, pendingId) };
    }
  }
}

// No unfinished UI offer: skip when the node is already on this release
// (type+version match and CRC matches the running-slot file). Stops A↔B ping-pong.
if (+request.type === +fw.type && +request.version === +fw.version && +request.type !== 0xFFFF &&
    !runFw.error && runFw.crc !== undefined && +request.crc === +runFw.crc) {
  return { noUpdate: true, labels: labels };
}

return { firmwareId: pick.firmwareId, blockSize: bs, labels: noteServe(labels, fw, pick.firmwareId) };
```

## data.onBlock

Leave empty. Core serves the `firmwareId` cached from `onConfig`.

## Future bootloaders

Create another data repository (e.g. `ota_future_bl`) with its own `onConfig`,
point `fota_script` at that id, and use whatever node labels that script documents.
No MyController code change.
