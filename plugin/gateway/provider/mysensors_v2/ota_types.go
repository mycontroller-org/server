package mysensors

import "time"

// firmwareBlockSize for more detail
// https://github.com/mysensors/MySensors/blob/2.3.2/core/MyOTAFirmwareUpdate.h#L68~L71
const firmwareBlockSize = 16

// firmwareConfigRequest data
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

// firmwareResponse data
type firmwareResponse struct {
	Type    uint16
	Version uint16
	Block   uint16
	Data    [firmwareBlockSize]uint8
}

// firmwareRaw returns firmwareRaw details
type firmwareRaw struct {
	Type       uint16    `json:"type" yaml:"type"`
	Version    uint16    `json:"version" yaml:"version"`
	Data       []uint8   `json:"data" yaml:"data"`
	Blocks     uint16    `json:"blocks" yaml:"blocks"`
	CRC        uint16    `json:"crc" yaml:"crc"`
	LastAccess time.Time `json:"lastAccess" yaml:"lastAccess"`
}
