package bus

import "github.com/mycontroller-org/backend/v2/pkg/model/event"

// bus client types
const (
	TypeEmbedded = "embedded"
	TypeNatsIO   = "nats_io"
)

// CallBackFunc message passed to this func
type CallBackFunc func(event *event.Event)

// Client interface
type Client interface {
	Close() error
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler CallBackFunc) (int64, error)
	Unsubscribe(topic string, subscriptionID int64) error
	UnsubscribeAll(topic string) error
}
