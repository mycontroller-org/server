package mysensors

// firmwareBlockSize for more detail
// https://github.com/mysensors/MySensors/blob/2.3.2/core/MyOTAFirmwareUpdate.h#L68~L71
const firmwareBlockSize = uint8(16)

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
//         https://github.com/mysensors/MySensorsBootloaderRF24/blob/3ed805edc18edd44db427c99f4ff53f6dcdbf502/MySensorsBootloader.h#L249
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
	Data    []uint8
}

// firmware returns firmware details
type firmware struct {
	Data   []byte
	Blocks int
	CRC    int
}
