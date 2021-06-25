package bus

import busML "github.com/mycontroller-org/server/v2/pkg/model/bus"

// bus client types
const (
	TypeEmbedded = "embedded"
	TypeNatsIO   = "natsio"
)

// CallBackFunc message passed to this func
type CallBackFunc func(data *busML.BusData)

// Client interface
type Client interface {
	Close() error
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler CallBackFunc) (int64, error)
	Unsubscribe(topic string, subscriptionID int64) error
	UnsubscribeAll(topic string) error
}
