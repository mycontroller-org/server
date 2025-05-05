package javascript_helper

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
)

type Convert struct {
}

// returns anything to string
func (ll *Convert) ToString(data interface{}) string {
	return convertor.ToString(data)
}

// returns string to byte
func (ll *Convert) ToBytes(data string) []byte {
	return []byte(data)
}

// returns anything to hex string
func (ll *Convert) ToHexString(data interface{}) string {
	return fmt.Sprintf("%X", data)
}

// returns string from hex string
func (ll *Convert) HexStringToString(data string) string {
	return string(ll.HexStringToBytes(data))
}

// returns bytes from hex string
func (ll *Convert) HexStringToBytes(data string) []byte {
	bytes, err := hex.DecodeString(data)
	if err != nil {
		return []byte(err.Error())
	}
	return bytes
}

// returns string from base64
func (ll *Convert) ToStringFromBase64(data interface{}) string {
	base64Str := convertor.ToString(data)
	bytes, err := base64.RawStdEncoding.DecodeString(base64Str)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

// ToUInt16LE converts a 2-byte little-endian slice to a uint16.
func (ll *Convert) ToUInt16LE(bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(bytes)
}

// ToUInt32LE converts a 4-byte little-endian slice to a uint32.
func (ll *Convert) ToUInt32LE(bytes []byte) uint32 {
	return binary.LittleEndian.Uint32(bytes)
}

// ToUInt64LE converts an 8-byte little-endian slice to a uint64.
func (ll *Convert) ToUInt64LE(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}

// ToInt16LE converts a 2-byte little-endian slice to an int16,
// preserving the sign using two's complement representation.
func (ll *Convert) ToInt16LE(bytes []byte) int16 {
	return int16(binary.LittleEndian.Uint16(bytes))
}

// ToInt32LE converts a 4-byte little-endian slice to an int32,
// preserving the sign using two's complement representation.
func (ll *Convert) ToInt32LE(bytes []byte) int32 {
	return int32(binary.LittleEndian.Uint32(bytes))
}

// ToInt64LE converts an 8-byte little-endian slice to an int64,
// preserving the sign using two's complement representation.
func (ll *Convert) ToInt64LE(bytes []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bytes))
}

// ToUInt16BE converts a 2-byte big-endian slice to a uint16.
func (ll *Convert) ToUInt16BE(bytes []byte) uint16 {
	return binary.BigEndian.Uint16(bytes)
}

// ToUInt32BE converts a 4-byte big-endian slice to a uint32.
func (ll *Convert) ToUInt32BE(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(bytes)
}

// ToUInt64BE converts an 8-byte big-endian slice to a uint64.
func (ll *Convert) ToUInt64BE(bytes []byte) uint64 {
	return binary.BigEndian.Uint64(bytes)
}

// ToInt16BE converts a 2-byte big-endian slice to an int16,
// preserving the sign using two's complement representation.
func (ll *Convert) ToInt16BE(bytes []byte) int16 {
	return int16(binary.BigEndian.Uint16(bytes))
}

// ToInt32BE converts a 4-byte big-endian slice to an int32,
// preserving the sign using two's complement representation.
func (ll *Convert) ToInt32BE(bytes []byte) int32 {
	return int32(binary.BigEndian.Uint32(bytes))
}

// ToInt64BE converts an 8-byte big-endian slice to an int64,
// preserving the sign using two's complement representation.
func (ll *Convert) ToInt64BE(bytes []byte) int64 {
	return int64(binary.BigEndian.Uint64(bytes))
}
