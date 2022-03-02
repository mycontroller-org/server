package javascript_helper

import (
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
