package mysensors

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

// FirmwareBlockSize for more detail
// https://github.com/mysensors/MySensors/blob/2.3.2/core/MyOTAFirmwareUpdate.h#L68~L71
const FirmwareBlockSize = uint8(8)

// FirmwareConfigRequest data
type FirmwareConfigRequest struct {
	Type      uint16
	Version   uint16
	Blocks    uint16
	Crc       uint16
	BLVersion uint16
}

// FirmwareConfigResponse data
type FirmwareConfigResponse struct {
	Type    uint16
	Version uint16
	Blocks  uint16
	Crc     uint16
}

// FirmwareRequest data
type FirmwareRequest struct {
	Type    uint16
	Version uint16
	Block   uint16
}

// FirmwareResponse data
type FirmwareResponse struct {
	Type    uint16
	Version uint16
	Block   uint16
	Data    []uint8
}

// ToHex returns hex string
func ToHex(data interface{}) (string, error) {
	var bBuf bytes.Buffer
	err := binary.Write(&bBuf, binary.LittleEndian, data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bBuf.Bytes()), nil
}

// ToStruct updates struct from hex string
func ToStruct(st interface{}, h string) error {
	hb, err := hex.DecodeString(h)
	if err != nil {
		return err
	}
	r := bytes.NewReader(hb)
	binary.Read(r, binary.LittleEndian, st)
	return nil
}
