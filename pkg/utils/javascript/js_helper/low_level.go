package javascript_helper

import "github.com/mycontroller-org/server/v2/pkg/utils/convertor"

type LowLevel struct {
}

// returns anything to string
func (ll *LowLevel) ToString(data interface{}) string {
	return convertor.ToString(data)
}

// returns string to byte
func (ll *LowLevel) ToByte(data string) []byte {
	return []byte(data)
}
