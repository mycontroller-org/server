package bus

// Client interface
type Client interface {
	Close() error
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler CallBackFunc) (int64, error)
	Unsubscribe(topic string, subscriptionID int64) error
	UnsubscribeAll(topic string) error
}

// CallBackFunc message passed to this func
type CallBackFunc func(event *Event)

// Event struct
type Event struct {
	Data interface{}
}

// bus client types
const (
	TypeEmbedded = "embedded"
	TypeNatsIO   = "nats_io"
)
