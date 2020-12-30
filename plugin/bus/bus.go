package bus

import "github.com/mycontroller-org/backend/v2/pkg/utils"

// Client interface
type Client interface {
	Close() error
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler CallBackFunc) (int64, error)
	Unsubscribe(topic string, subscriptionID int64) error
	UnsubscribeAll(topic string) error
}

// bus client types
const (
	TypeEmbedded = "embedded"
	TypeNatsIO   = "nats_io"
)

// CallBackFunc message passed to this func
type CallBackFunc func(event *Event)

// Event struct
type Event struct {
	Data []byte
}

// SetData updates data in []byte format
func (e *Event) SetData(data interface{}) error {
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
func (e *Event) ToStruct(out interface{}) error {
	return utils.ByteToStruct(e.Data, out)
}
