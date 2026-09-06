package mysensors

import "time"

// defaultFirmwareBlockSize is used only for stock DualOptiboot nodes that never
// report protocol 3.1 blockSize. A node with fota_script must not fall back to 16
// unless it actually advertised 16.
// https://github.com/mysensors/MySensors/blob/2.3.2/core/MyOTAFirmwareUpdate.h#L68~L71
const defaultFirmwareBlockSize = 16

// maxFirmwareBlockSize is the largest OTA data slice this controller will serve.
// The node/gateway fork extends V2 length (5-bit field = 31 means "see _extLength").
// LoRa 255-byte packets + AES pad + 6-byte radio header comfortably fit 192.
// Raise this if the radio MTU and MQTT_MAX_PACKET_SIZE are increased together.
const maxFirmwareBlockSize = 192

// LabelOtaBlockSize is set from the node's CONFIG_REQUEST blockSize (protocol 3.1).
const LabelOtaBlockSize = "ms_ota_block_size"

// fotaFirmwareHeadBytes is how many leading image bytes getFirmware() exposes
// to FOTA scripts as `head` (hex). Scripts interpret the contents.
const fotaFirmwareHeadBytes = 16

// firmwareConfigRequest data (stock 10-byte layout; longer payloads still decode first 10)
type firmwareConfigRequest struct {
	Type              uint16
	Version           uint16
	Blocks            uint16
	CRC               uint16
	BootloaderVersion uint16
}

// firmwareConfigResponse data
type firmwareConfigResponse struct {
	Type    uint16
	Version uint16
	Blocks  uint16
	CRC     uint16
}

// SetEraseEEPROM Erases EEPROM of a node
// Source: https://github.com/mysensors/MySensorsBootloaderRF24/blob/3ed805edc18edd44db427c99f4ff53f6dcdbf502/MySensorsBootloader.h#L29
//
//	https://github.com/mysensors/MySensorsBootloaderRF24/blob/3ed805edc18edd44db427c99f4ff53f6dcdbf502/MySensorsBootloader.h#L249
func (fwCres *firmwareConfigResponse) SetEraseEEPROM() {
	fwCres.Blocks = 0
	fwCres.CRC = 0xDA7A
	fwCres.Type = 0x01
}

// firmwareRequest data
type firmwareRequest struct {
	Type    uint16
	Version uint16
	Block   uint16
}

// firmwareResponse is the stock 16-byte-data ST_FIRMWARE_RESPONSE (6 + 16 = 22).
// Same layout the 328 + RFM69 path used via toHex.
type firmwareResponse struct {
	Type    uint16
	Version uint16
	Block   uint16
	Data    [defaultFirmwareBlockSize]uint8
}

// firmwareRaw is a contiguous OTA image sliced to BlockSize.
type firmwareRaw struct {
	Type       uint16    `json:"type" yaml:"type"`
	Version    uint16    `json:"version" yaml:"version"`
	Data       []uint8   `json:"data" yaml:"data"`
	Blocks     uint16    `json:"blocks" yaml:"blocks"`
	CRC        uint16    `json:"crc" yaml:"crc"`
	BlockSize  int       `json:"blockSize" yaml:"blockSize"`
	LastAccess time.Time `json:"lastAccess" yaml:"lastAccess"`
}
