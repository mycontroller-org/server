package event

import "github.com/mycontroller-org/backend/v2/pkg/utils"

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
	bytes, err := utils.StructToByte(data)
	if err != nil {
		return err
	}
	e.Data = bytes
	return nil
}

// ToStruct converts data to target interface
func (e *BusData) ToStruct(out interface{}) error {
	return utils.ByteToStruct(e.Data, out)
}
