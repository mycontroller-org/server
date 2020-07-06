package serial

import (
	q "github.com/jaegertracing/jaeger/pkg/queue"
	m2s "github.com/mitchellh/mapstructure"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	s "github.com/tarm/serial"
)

// Constants in serial gateway
const (
	KeyMessageSplitter = "MessageSplitter"
)

// Config details
type Config struct {
	Portname       string
	BaudRate       int
	MessageSpliter byte
}

// Endpoint data
type Endpoint struct {
	Config    Config
	Port      *s.Port
	TxQueue   *q.BoundedQueue
	RxQueue   *q.BoundedQueue
	GatewayID string
}

// New serial client
func New(config map[string]interface{}, txQueue, rxQueue *q.BoundedQueue, gID string) (*Endpoint, error) {
	var cfg Config
	err := m2s.Decode(config, &cfg)
	if err != nil {
		return nil, err
	}
	c := &s.Config{Name: cfg.Portname, Baud: cfg.BaudRate}
	port, err := s.OpenPort(c)
	if err != nil {
		return nil, err
	}
	d := &Endpoint{
		Config:    cfg,
		Port:      port,
		TxQueue:   txQueue,
		RxQueue:   rxQueue,
		GatewayID: gID,
	}
	return d, nil
}

func (d *Endpoint) Write(rm *msg.RawMessage) error {
	_, err := d.Port.Write(rm.Data)
	return err
}

// Close the driver
func (d *Endpoint) Close() error {
	err := d.Port.Close()
	d.TxQueue.Stop()
	d.RxQueue.Stop()
	return err
}
