package javascript_helper

import (
	"encoding/base64"
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

// returns string from base64
func (ll *Convert) ToStringFromBase64(data interface{}) string {
	base64Str := convertor.ToString(data)
	bytes, err := base64.RawStdEncoding.DecodeString(base64Str)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}
