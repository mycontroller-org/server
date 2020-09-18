package tasmota

import "encoding/json"

func toStruct(data []byte, out interface{}) error {
	return json.Unmarshal(data, out)
}
