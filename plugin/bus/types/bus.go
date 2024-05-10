package bus

import (
	"context"
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
)

const (
	contextKey types.ContextKey = "bus_plugin"
)

// Plugin interface
type Plugin interface {
	Name() string
	Close() error
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler CallBackFunc) (int64, error)
	Unsubscribe(topic string, subscriptionID int64) error
	QueueSubscribe(topic, queueName string, handler CallBackFunc) (int64, error)
	QueueUnsubscribe(topic, queueName string, subscriptionID int64) error
	UnsubscribeAll(topic string) error
	PausePublish()
	ResumePublish()
	TopicPrefix() string
}

func FromContext(ctx context.Context) (Plugin, error) {
	bus, ok := ctx.Value(contextKey).(Plugin)
	if !ok {
		return nil, errors.New("invalid bus instance received in context")
	}
	if bus == nil {
		return nil, errors.New("bus instance not provided in context")
	}
	return bus, nil
}

func WithContext(ctx context.Context, bus Plugin) context.Context {
	return context.WithValue(ctx, contextKey, bus)
}

// CallBackFunc message passed to this func
type CallBackFunc func(data *BusData)

// BusData struct
type BusData struct {
	Topic string `json:"topic" yaml:"topic"`
	Data  []byte `json:"data" yaml:"data"`
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
