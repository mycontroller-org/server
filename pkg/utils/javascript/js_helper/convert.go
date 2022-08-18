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

func (ll *Convert) ToUInt16LE(bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(bytes)
}

func (ll *Convert) ToInt16LE(bytes []byte) int16 {
	ref := ll.ToUInt16LE(bytes)
	if ref > 0x7fff {
		return int16(-65536 + int(ref))
	} else {
		return int16(ref)
	}
}

func (ll *Convert) ToUInt16BE(bytes []byte) uint16 {
	return binary.BigEndian.Uint16(bytes)
}
