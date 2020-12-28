package bus

// Client interface
type Client interface {
	Close() error
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler func(event *Event)) error
	Unsubscribe(topic string) error
}

// Event struct
type Event struct {
	Data interface{}
}

// bus client types
const (
	TypeEmbedded = "embedded"
	TypeNatsIO   = "nats_io"
)
