package event

import (
	"github.com/mycontroller-org/backend/v2/pkg/json"
)

// BusData struct
type BusData struct {
	Topic string
	Data  []byte
}

// SetData updates data in []byte format
func (e *BusData) SetData(data interface{}) error {
	if data == nil {
		return nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	e.Data = bytes
	return nil
}

// LoadData converts data to target interface
func (e *BusData) LoadData(out interface{}) error {
	err := json.Unmarshal(e.Data, out)
	return err
}
